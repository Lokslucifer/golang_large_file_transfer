package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
}

type Transfer struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	OwnerID      uuid.UUID  `json:"owner_id" db:"owner_id"`
	TransferPath string     `json:"transfer_path" db:"transfer_path"`
	Size         int64      `json:"size" db:"size"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	Expiry       *time.Time `json:"expiry" db:"expiry"`
	Message      string     `json:"message" db:"message"`
}

type File struct {
	ID                uuid.UUID `json:"id" db:"id"`
	FileName          string    `json:"file_name" db:"file_name"`
	FileSize          int64     `json:"file_size" db:"file_size"`
	FilePath          string    `json:"file_path" db:"file_path"`
	TransferID        uuid.UUID `json:"transfer_id" db:"transfer_id"`
	FileExtension     string    `json:"file_extension" db:"file_extension"`
	NumOfActiveStream int       `json:"num_of_active_stream" db:"num_of_active_stream"`
}

type TempTransfer struct {
	ID          uuid.UUID `json:"id" db:"id"`
	OwnerID     uuid.UUID `json:"owner_id" db:"owner_id"`
	Message     string    `json:"message" db:"message"`
	Size        int64     `json:"size" db:"size"`
	Expiry      string    `json:"expiry" db:"expiry"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	LastUpdated time.Time `json:"last_updated" db:"last_updated"`
}

type Chunk struct {
	ID         uuid.UUID `json:"id" db:"id"`
	TranferID  uuid.UUID `json:"transfer_id" db:"transfer_id"`
	Index      int       `json:"index" db:"index"`
	UploadedAt time.Time `json:"uploaded_at" db:"uploaded_at"`
}

type SysFileInfo struct {
	Name    string    `json:"name"`     // Base name of the file
	Size    int64     `json:"size"`     // Length in bytes
	ModTime time.Time `json:"mod_time"` // Last modification time
	IsDir   bool      `json:"is_dir"`   // Is it a directory
	Path    string    `json:"path"`
}

// type Link struct {
// 	ID          uuid.UUID `json:"id" db:"id"`                 // UUID
// 	FileID      uuid.UUID `json:"file_id" db:"file_id"`       // FK to File.ID
// 	Type        string    `json:"type" db:"type"`             // "public" or "private"
// 	AccessList  []string  `json:"access_list" db:"-"`         // Not directly DB-storable
// 	CreatedByID string    `json:"created_by" db:"created_by"` // FK to User.ID
// 	ExpiresAt   string    `json:"expires_at" db:"expires_at"` // Optional expiration
// }

// type LinkAccess struct {
// 	LinkID uuid.UUID `db:"link_id"`
// 	UserID uuid.UUID `db:"user_id"`
// }
