package repository

import (
	"context"
	"fmt"
	"large_fss/internals/constants"
	"large_fss/internals/models"
	"time"

	"github.com/google/uuid"
)

// CreateTempTransfer inserts a new temp transfer and returns its ID.
func (p *PostgresSQLDB) CreateTempTransfer(ctx context.Context, temptrans models.TempTransfer) (uuid.UUID, error) {
	temptrans.ID = uuid.New()
	temptrans.CreatedAt = time.Now()
	temptrans.LastUpdated = time.Now()

	query := `
		INSERT INTO temp_transfers (id, owner_id, message,size, expiry, created_at, last_updated)
		VALUES (:id, :owner_id,:message, :size, :expiry, :created_at, :last_updated)`

	_, err := p.db.NamedExecContext(ctx, query, &temptrans)
	if err != nil {

		return uuid.Nil, fmt.Errorf("postgres: create temp transfer: %w", err)
	}

	return temptrans.ID, nil
}

func (p *PostgresSQLDB) FindAllFailedTempTransfers(ctx context.Context) ([]models.TempTransfer, error) {
	query := fmt.Sprintf(`
		SELECT id, owner_id, message, size, expiry, created_at, last_updated
		FROM temp_transfers
		WHERE last_updated < NOW() - INTERVAL '%d hours'
		ORDER BY last_updated ASC;
	`, constants.MaxhoursUploadSessionValid)

	rows, err := p.db.QueryxContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("postgres: query failed temp transfers older than %d hours: %w", constants.MaxhoursUploadSessionValid, err)
	}
	defer rows.Close()

	var failedTransfers []models.TempTransfer
	for rows.Next() {
		var t models.TempTransfer
		if err := rows.StructScan(&t); err != nil {
			return nil, fmt.Errorf("postgres: scan failed: %w", err)
		}
		failedTransfers = append(failedTransfers, t)
	}

	return failedTransfers, nil
}

func (p *PostgresSQLDB) DeleteTempTransferByID(ctx context.Context, transferID uuid.UUID) error {
	query := `DELETE FROM temp_transfers WHERE id = $1`

	_, err := p.db.ExecContext(ctx, query, transferID)
	if err != nil {
		return fmt.Errorf("postgres: delete temp transfer by ID %s: %w", transferID, err)
	}
	return nil
}

// FindTempTransferByID fetches a temp transfer by its ID.
func (p *PostgresSQLDB) FindTempTransferByID(ctx context.Context, id uuid.UUID) (*models.TempTransfer, error) {
	var temptrans models.TempTransfer
	query := `SELECT * FROM temp_transfers WHERE id = $1`

	err := p.db.GetContext(ctx, &temptrans, query, id)
	if err != nil {
		return nil, fmt.Errorf("postgres: find temp transfer by ID %s: %w", id, err)
	}

	return &temptrans, nil
}

// UpdateTempTransferLastUpdatedTimeByID updates the last_updated timestamp.
func (p *PostgresSQLDB) UpdateTempTransferLastUpdatedTimeByID(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE temp_transfers SET last_updated = $1 WHERE id = $2`

	_, err := p.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("postgres: update temp transfer last updated time by ID %s: %w", id, err)
	}
	return nil
}

// CreateTransfer inserts a new permanent transfer record.
func (p *PostgresSQLDB) CreateTransfer(ctx context.Context, trans models.Transfer) (uuid.UUID, error) {
	fmt.Println("transfer-", trans)
	trans.CreatedAt = time.Now()

	query := `
		INSERT INTO transfers (id, owner_id, transfer_path,message, size, created_at, expiry)
		VALUES (:id, :owner_id, :transfer_path,:message, :size, :created_at, :expiry)`

	_, err := p.db.NamedExecContext(ctx, query, &trans)
	if err != nil {
		return uuid.Nil, fmt.Errorf("postgres: create Transfer: %w", err)
	}

	return trans.ID, nil
}

// Update Transfer
func (p *PostgresSQLDB) UpdateTransferByID(ctx context.Context, trans models.Transfer) error {
	query := `
		UPDATE transfers SET message = $1,expiry = $2
		WHERE id = $3 `
	_, err := p.db.ExecContext(ctx, query, trans.Message, trans.Expiry, trans.ID)
	if err != nil {
		return fmt.Errorf("postgres: update transfer by id %v: %w", trans.ID, err)
	}

	return nil
}

// Delete Transfer
func (p *PostgresSQLDB) DeleteTransferByID(ctx context.Context, transferID uuid.UUID) error {
	query := `DELETE FROM transfers WHERE id = $1`

	_, err := p.db.ExecContext(ctx, query, transferID)
	if err != nil {
		return fmt.Errorf("postgres: delete transfer by ID %s: %w", transferID, err)
	}
	return nil
}

// CreateFile inserts a file associated with a transfer.
func (p *PostgresSQLDB) CreateFile(ctx context.Context, fileData models.File) (uuid.UUID, error) {
	fileData.ID = uuid.New()

	query := `
		INSERT INTO files (id, file_name, file_size, file_path, transfer_id, file_extension)
		VALUES (:id, :file_name, :file_size, :file_path, :transfer_id, :file_extension)`

	_, err := p.db.NamedExecContext(ctx, query, &fileData)
	if err != nil {
		return uuid.Nil, fmt.Errorf("postgres: create file: %w", err)
	}

	return fileData.ID, nil
}

func (p *PostgresSQLDB) CreateChunk(ctx context.Context, chunk models.Chunk) error {
	chunk.ID = uuid.New()
	query := `
		INSERT INTO chunks (id,transfer_id, index, uploaded_at)
		VALUES ($1, $2, $3,$4)`
	_, err := p.db.ExecContext(ctx, query, chunk.ID, chunk.TranferID, chunk.Index, chunk.UploadedAt)
	if err != nil {
		return fmt.Errorf("postgres: create chunk: %w", err)
	}
	return nil
}

func (p *PostgresSQLDB) GetAllUploadedChunksIndex(ctx context.Context, transferID uuid.UUID) ([]int, error) {
	query := `SELECT index FROM chunks WHERE transfer_id = $1 ORDER BY index ASC`
	rows, err := p.db.QueryxContext(ctx, query, transferID)
	if err != nil {
		return nil, fmt.Errorf("postgres: get all upload chunk index by transfer ID %s: %w", transferID, err)
	}
	defer rows.Close()

	var indexes []int
	for rows.Next() {
		var index int
		if err := rows.Scan(&index); err != nil {
			return nil, err
		}
		indexes = append(indexes, index)
	}
	return indexes, nil
}

func (p *PostgresSQLDB) FindAllFilesByTransferID(ctx context.Context, transferID uuid.UUID) ([]models.File, error) {
	query := `SELECT * FROM files WHERE transfer_id = $1`
	var files []models.File
	err := p.db.SelectContext(ctx, &files, query, transferID)
	if err != nil {
		return nil, fmt.Errorf("postgres: find all files by TransferID %s: %w", transferID, err)
	}
	return files, nil
}

func (p *PostgresSQLDB) FindFileByID(ctx context.Context, fileID uuid.UUID) (*models.File, error) {
	query := `SELECT * FROM files WHERE id = $1`
	var file models.File
	err := p.db.GetContext(ctx, &file, query, fileID)
	if err != nil {
		return nil, fmt.Errorf("postgres: find file by FileID %s: %w", fileID, err)
	}
	return &file, nil
}

func (p *PostgresSQLDB) IncrementActiveStreamByID(ctx context.Context, fileID uuid.UUID) error {
	query := `
        UPDATE files 
        SET num_of_active_stream = num_of_active_stream + 1 
        WHERE id = $1;
    `
	result, err := p.db.ExecContext(ctx, query, fileID)
	if err != nil {
		return fmt.Errorf("failed to increment active stream: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("no file found with ID %s", fileID)
	}
	return nil
}

func (p *PostgresSQLDB) DecrementActiveStreamByID(ctx context.Context, fileID uuid.UUID) error {
	query := `
        UPDATE files 
        SET num_of_active_stream = GREATEST(num_of_active_stream - 1, 0)
        WHERE id = $1;
    `
	result, err := p.db.ExecContext(ctx, query, fileID)
	if err != nil {
		return fmt.Errorf("failed to decrement active stream: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("no file found with ID %s", fileID)
	}
	return nil
}
func (p *PostgresSQLDB) FindTransferByID(ctx context.Context, transferID uuid.UUID) (*models.Transfer, error) {
	query := `SELECT * FROM transfers WHERE id = $1`
	var transfer models.Transfer
	err := p.db.GetContext(ctx, &transfer, query, transferID)
	if err != nil {
		return nil, fmt.Errorf("postgres: find transfer by TransferID %s: %w", transferID, err)
	}
	return &transfer, nil
}

func (p *PostgresSQLDB) FindAllTransfersByUserID(ctx context.Context, userID uuid.UUID) ([]models.Transfer, error) {
	query := `SELECT * FROM transfers WHERE owner_id = $1 ORDER BY created_at DESC`
	var transfers []models.Transfer
	err := p.db.SelectContext(ctx, &transfers, query, userID)
	if err != nil {
		return nil, fmt.Errorf("postgres: find all transfers by UserID %s: %w", userID, err)
	}
	return transfers, nil
}

func (p *PostgresSQLDB) FindAllExpiredTransfers(ctx context.Context) ([]models.Transfer, error) {
	query := `
		SELECT id, owner_id, transfer_path, message, size, created_at, expiry
		FROM transfers
		WHERE expiry IS NOT NULL AND expiry < NOW()
		ORDER BY expiry ASC;
	`

	rows, err := p.db.QueryxContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("postgres: get all expired transfers: %w", err)
	}
	defer rows.Close()

	var transfers []models.Transfer
	for rows.Next() {
		var t models.Transfer
		if err := rows.StructScan(&t); err != nil {
			return nil, fmt.Errorf("postgres: scanning expired transfer: %w", err)
		}
		transfers = append(transfers, t)
	}

	return transfers, nil
}

// func (p *PostgresSQLDB)CreateChunk(ctx context.Context,chunk models.Chunk)(error){
// 	return nil
// }

// func (p *PostgresSQLDB)GetAllUploadedChunksIndex(ctx context.Context,transferID uuid.UUID)([]int,error){
// 		return []int{},nil
// 	}

// func (p *PostgresSQLDB)FindAllFilesByTransferID(ctx context.Context,transferID uuid.UUID)([]models.File,error){
// 	return []models.File{},nil
// }

// func (p *PostgresSQLDB)FindFileByID(ctx context.Context,fileID uuid.UUID)(*models.File,error){
// 	return nil,nil
// }

// func (p *PostgresSQLDB)FindTransferByID(ctx context.Context,transferID uuid.UUID)(*models.Transfer,error){
// 	return nil,nil
// }
