package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "graphweaver"
	}
	dbPass := os.Getenv("DB_PASSWORD")
	if dbPass == "" {
		dbPass = "graphweaver123"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "graphweaver"
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)
	
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Migration SQL
	schema := `
	CREATE TABLE IF NOT EXISTS notebooks (
		id BIGSERIAL PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		description TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	ALTER TABLE documents ADD COLUMN IF NOT EXISTS notebook_id BIGINT REFERENCES notebooks(id) ON DELETE CASCADE;
	ALTER TABLE documents ADD COLUMN IF NOT EXISTS summary TEXT;
	ALTER TABLE documents ADD COLUMN IF NOT EXISTS error_message TEXT;

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
	`

	_, err = db.Exec(schema)
	if err != nil {
		log.Fatalf("Failed to execute migration: %v", err)
	}

	fmt.Println("Migration executed successfully.")
}
