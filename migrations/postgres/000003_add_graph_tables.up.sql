CREATE TABLE IF NOT EXISTS nodes (
    id BIGSERIAL PRIMARY KEY,
    document_id BIGINT REFERENCES documents(id) ON DELETE CASCADE,
    label VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    properties JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS edges (
    id BIGSERIAL PRIMARY KEY,
    document_id BIGINT REFERENCES documents(id) ON DELETE CASCADE,
    source_node_id BIGINT REFERENCES nodes(id) ON DELETE CASCADE,
    target_node_id BIGINT REFERENCES nodes(id) ON DELETE CASCADE,
    relation_type VARCHAR(255) NOT NULL,
    properties JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_nodes_document_id ON nodes(document_id);
CREATE INDEX idx_edges_document_id ON edges(document_id);
CREATE INDEX idx_edges_source_target ON edges(source_node_id, target_node_id);
