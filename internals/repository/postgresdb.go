package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type PostgresSQLDB struct {
	db *sqlx.DB
}

func NewPostgresSQLDB(db *sqlx.DB) DbRepository {
	createTables(db)
	return &PostgresSQLDB{db: db}
}

// CreateTables creates the necessary database tables.
func createTables(db *sqlx.DB) {
	fmt.Println("Creating tables...")

	// Install uuid-ossp extension
	extensionQuery := `CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`
	if _, err := db.Exec(extensionQuery); err != nil {
		fmt.Printf("Error in installing extension uuid-ossp: %v\n", err)
	}

	// Start transaction
	tx, err := db.Beginx()
	if err != nil {
		fmt.Printf("Error starting transaction: %v\n", err)
		return
	}

	// Helper to execute table creation queries within the transaction
	executeTableQuery := func(query, tableName string) {
		if _, err := tx.Exec(query); err != nil {
			fmt.Printf("Error creating %s table: %v\n", tableName, err)
		} else {
			fmt.Printf("%s table created successfully.\n", tableName)
		}
	}

	// users table
	userTableQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		first_name TEXT,
		last_name TEXT
	);`
	executeTableQuery(userTableQuery, "users")

	// transfers table
	transferTableQuery := `
	CREATE TABLE IF NOT EXISTS transfers (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		owner_id UUID NOT NULL,
		transfer_path TEXT NOT NULL,
		message TEXT DEFAULT '',
		size BIGINT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		expiry TIMESTAMP WITH TIME ZONE,
		FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
	);`
	executeTableQuery(transferTableQuery, "transfers")

	// files table
	fileTableQuery := `
	CREATE TABLE IF NOT EXISTS files (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		file_name TEXT NOT NULL,
		file_size BIGINT NOT NULL,
		file_path TEXT NOT NULL,
		transfer_id UUID NOT NULL,
		file_extension TEXT,
		num_of_active_stream  INT DEFAULT 0,
		FOREIGN KEY (transfer_id) REFERENCES transfers(id) ON DELETE CASCADE
	);`
	executeTableQuery(fileTableQuery, "files")

	// temp_transfers table
	tempTransferTableQuery := `
	CREATE TABLE IF NOT EXISTS temp_transfers (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		owner_id UUID NOT NULL,
		message TEXT DEFAULT '',
		size BIGINT NOT NULL,
		expiry TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		last_updated TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
	);`
	executeTableQuery(tempTransferTableQuery, "temp_transfers")

	chunkTableQuery := `
	CREATE TABLE IF NOT EXISTS chunks (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		transfer_id UUID NOT NULL,
		index INTEGER NOT NULL,
		uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		FOREIGN KEY (transfer_id) REFERENCES temp_transfers(id) ON DELETE CASCADE
	);`
	executeTableQuery(chunkTableQuery, "chunks")

	// Commit transaction
	if err := tx.Commit(); err != nil {
		fmt.Printf("Error committing transaction: %v\n", err)
		if rberr := tx.Rollback(); rberr != nil {
			fmt.Printf("Error rolling back transaction: %v\n", rberr)
		}
	}
}
