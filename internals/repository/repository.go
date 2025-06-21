package repository

import (
	"context"
	"large_fss/internals/models"

	"github.com/google/uuid"
)

type DbRepository interface {

	//Users
	CreateUser(ctx context.Context, user models.User) (uuid.UUID, error)

	FindAllUsers(ctx context.Context) ([]models.User, error)

	// Retrieves a user by their email
	FindUserByEmail(ctx context.Context, email string) (*models.User, error)

	// Retrieves a user by their ID
	FindUserById(ctx context.Context, id uuid.UUID) (*models.User, error)

	//Resources
	CreateTempTransfer(ctx context.Context, temptrans models.TempTransfer) (uuid.UUID, error)

	CreateChunk(ctx context.Context,chunk models.Chunk)(error)

	GetAllUploadedChunksIndex(ctx context.Context,transferID uuid.UUID)([]int,error)

	DeleteTempTransferByID(ctx context.Context,transferID uuid.UUID)(error)

	FindAllFailedTempTransfers(ctx context.Context)([]models.TempTransfer,error)


	FindTempTransferByID(ctx context.Context, id uuid.UUID) (*models.TempTransfer, error)

	UpdateTempTransferLastUpdatedTimeByID(ctx context.Context, id uuid.UUID)(error)

	CreateTransfer(ctx context.Context, trans models.Transfer) (uuid.UUID, error)

	UpdateTransferByID(ctx context.Context,trans models.Transfer)(error)

	DeleteTransferByID(ctx context.Context,transID uuid.UUID)(error)

	FindAllExpiredTransfers(ctx context.Context)([]models.Transfer,error)

	





	CreateFile(ctx context.Context, fileData models.File)(uuid.UUID,error)

	FindTransferByID(ctx context.Context,transferID uuid.UUID)(*models.Transfer,error)

	FindAllFilesByTransferID(ctx context.Context,transferID uuid.UUID)([]models.File,error)

	FindFileByID(ctx context.Context,fileID uuid.UUID)(*models.File,error)

	IncrementActiveStreamByID(ctx context.Context,fileID uuid.UUID)(error)
	DecrementActiveStreamByID(ctx context.Context,fileID uuid.UUID)(error)

	FindAllTransfersByUserID(ctx context.Context,userID uuid.UUID)([]models.Transfer,error)


	// ModifyTimeById(ctx context.Context,id uuid.UUID)(error)

	// GetFileById(ctx context.Context, id uuid.UUID) (*models.FileMetaData, error)

	// UpdateFileName(ctx context.Context, fileId uuid.UUID, fileName string) error

	// UpdateFileDeleteStatus(c context.Context, fileId uuid.UUID) error

	// DeleteFileById(c context.Context, fileId uuid.UUID) error


}

