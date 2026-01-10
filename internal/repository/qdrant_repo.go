package repository

import (
	"context"
	"fmt"

	pb "github.com/qdrant/go-client/qdrant"
	"github.com/suyw-0123/graphweaver/internal/entity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// QdrantVectorRepository implements VectorRepository using Qdrant
type QdrantVectorRepository struct {
	conn              *grpc.ClientConn
	collectionsClient pb.CollectionsClient
	pointsClient      pb.PointsClient
}

// NewQdrantVectorRepository creates a new Qdrant repository
func NewQdrantVectorRepository(host string, port int) (*QdrantVectorRepository, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to qdrant: %w", err)
	}

	return &QdrantVectorRepository{
		conn:              conn,
		collectionsClient: pb.NewCollectionsClient(conn),
		pointsClient:      pb.NewPointsClient(conn),
	}, nil
}

// Close closes the gRPC connection
func (r *QdrantVectorRepository) Close() error {
	return r.conn.Close()
}

// CreateCollection creates a new collection
func (r *QdrantVectorRepository) CreateCollection(ctx context.Context, name string, vectorSize int) error {
	// Check if exists first
	exists, err := r.collectionsClient.CollectionExists(ctx, &pb.CollectionExistsRequest{
		CollectionName: name,
	})
	if err != nil {
		return fmt.Errorf("failed to check collection existence: %w", err)
	}
	if exists.Result != nil && exists.Result.Exists {
		return nil
	}

	_, err = r.collectionsClient.Create(ctx, &pb.CreateCollection{
		CollectionName: name,
		VectorsConfig: &pb.VectorsConfig{
			Config: &pb.VectorsConfig_Params{
				Params: &pb.VectorParams{
					Size:     uint64(vectorSize),
					Distance: pb.Distance_Cosine,
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}
	return nil
}

// DeleteCollection deletes a collection
func (r *QdrantVectorRepository) DeleteCollection(ctx context.Context, name string) error {
	_, err := r.collectionsClient.Delete(ctx, &pb.DeleteCollection{
		CollectionName: name,
	})
	if err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}
	return nil
}

// Upsert stores or updates vectors
func (r *QdrantVectorRepository) Upsert(ctx context.Context, collection string, points []*entity.VectorPoint) error {
	qPoints := make([]*pb.PointStruct, len(points))
	for i, p := range points {
		// Convert payload map to struct
		payload := make(map[string]*pb.Value)
		for k, v := range p.Payload {
			// Simple conversion for strings and numbers
			// For complex types, more robust conversion needed
			switch val := v.(type) {
			case string:
				payload[k] = &pb.Value{Kind: &pb.Value_StringValue{StringValue: val}}
			case int:
				payload[k] = &pb.Value{Kind: &pb.Value_IntegerValue{IntegerValue: int64(val)}}
			case int64:
				payload[k] = &pb.Value{Kind: &pb.Value_IntegerValue{IntegerValue: val}}
			case float64:
				payload[k] = &pb.Value{Kind: &pb.Value_DoubleValue{DoubleValue: val}}
			}
		}

		qPoints[i] = &pb.PointStruct{
			Id: &pb.PointId{
				PointIdOptions: &pb.PointId_Uuid{Uuid: p.ID},
			},
			Vectors: &pb.Vectors{
				VectorsOptions: &pb.Vectors_Vector{Vector: &pb.Vector{Data: p.Vector}},
			},
			Payload: payload,
		}
	}

	_, err := r.pointsClient.Upsert(ctx, &pb.UpsertPoints{
		CollectionName: collection,
		Points:         qPoints,
	})
	if err != nil {
		return fmt.Errorf("failed to upsert points: %w", err)
	}
	return nil
}

// Search find nearest neighbors
func (r *QdrantVectorRepository) Search(ctx context.Context, collection string, vector []float32, limit int, scoreThreshold float32) ([]SearchResult, error) {
	res, err := r.pointsClient.Search(ctx, &pb.SearchPoints{
		CollectionName: collection,
		Vector:         vector,
		Limit:          uint64(limit),
		ScoreThreshold: &scoreThreshold,
		WithPayload: &pb.WithPayloadSelector{
			SelectorOptions: &pb.WithPayloadSelector_Enable{Enable: true},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	results := make([]SearchResult, len(res.Result))
	for i, hit := range res.Result {
		// Convert payload back to map
		payload := make(map[string]interface{})
		for k, v := range hit.Payload {
			switch knd := v.Kind.(type) {
			case *pb.Value_StringValue:
				payload[k] = knd.StringValue
			case *pb.Value_IntegerValue:
				payload[k] = knd.IntegerValue
			case *pb.Value_DoubleValue:
				payload[k] = knd.DoubleValue
			}
		}

		id := ""
		if hit.Id != nil {
			if uuid := hit.Id.GetUuid(); uuid != "" {
				id = uuid
			}
		}

		results[i] = SearchResult{
			ID:      id,
			Score:   hit.Score,
			Payload: payload,
		}
	}
	return results, nil
}

// Delete removes points by ID
func (r *QdrantVectorRepository) Delete(ctx context.Context, collection string, ids []string) error {
	points := make([]*pb.PointId, len(ids))
	for i, id := range ids {
		points[i] = &pb.PointId{
			PointIdOptions: &pb.PointId_Uuid{Uuid: id},
		}
	}

	_, err := r.pointsClient.Delete(ctx, &pb.DeletePoints{
		CollectionName: collection,
		Points: &pb.PointsSelector{
			PointsSelectorOneOf: &pb.PointsSelector_Points{
				Points: &pb.PointsIdsList{Ids: points},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete points: %w", err)
	}
	return nil
}
