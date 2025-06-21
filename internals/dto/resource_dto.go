package dto

import (
	"mime/multipart"

	"time"

	"github.com/google/uuid"
)

type TransferDTO struct {
	Message string `json:"message"`
	Size    int64  `json:"size"`
	Expiry  string `json:"expiry"`
	OwnerID uuid.UUID
}

type CancelTransferDTO struct {
	ID      string `json:"transfer_id"`
	OwnerID uuid.UUID
}
type FileAssembleDTO struct {
	ID      uuid.UUID
	OwnerID uuid.UUID
}

type ChunkUploadDTO struct {
	ID         uuid.UUID
	ChunkIndex int
	OwnerID    uuid.UUID
	FileChunk  *multipart.FileHeader
}

type FileUpdateDTO struct {
	FileID   uuid.UUID
	FileName string `json:"file_name" db:"file_name"`
	OwnerID  uuid.UUID
}

type TransferInfoDTO struct {
	ID           uuid.UUID     `json:"id"`
	Message      string        `json:"message"`
	Size         int64         `json:"size"`
	Expiry       time.Time     `json:"expiry"`
	FileInfoList []FileInfoDTO `json:"file_info_list,omitempty"`
	CreatedAt   time.Time  `json:"created_at" `
}

type FileInfoDTO struct {
	ID            uuid.UUID `json:"id"`
	FileName      string    `json:"file_name" `
	FileSize      int64     `json:"file_size" `
	FileExtension string    `json:"file_extension" `
}

type TransferUpdateDTO struct {
	TransferID uuid.UUID `json:"transfer_id"`
	Message    string    `json:"message"`
	Expiry     string    `json:"expiry"`
	OwnerID    uuid.UUID

}


