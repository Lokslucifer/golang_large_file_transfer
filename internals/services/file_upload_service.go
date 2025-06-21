package services

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"large_fss/internals/constants"
	customerrors "large_fss/internals/customErrors"
	"large_fss/internals/dto"
	"large_fss/internals/models"
	"large_fss/utils"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *Service) CreateTransferService(c context.Context, fileUploadRequest dto.TransferDTO) (uuid.UUID, error) {
	var tempTransfer models.TempTransfer
	tempTransfer.Size = fileUploadRequest.Size
	tempTransfer.OwnerID = fileUploadRequest.OwnerID
	tempTransfer.CreatedAt = time.Now()
	tempTransfer.Expiry = fileUploadRequest.Expiry
	tempTransfer.Message = fileUploadRequest.Message
	fileId, err := s.repo.CreateTempTransfer(c, tempTransfer)
	if err != nil {
		return uuid.UUID{}, err
	}
	return fileId, nil

}

func (s *Service) CancelTransferService(c context.Context, transferID uuid.UUID, ownerID uuid.UUID) error {
	tempTransferData, err := s.repo.FindTempTransferByID(c, transferID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return customerrors.ErrUploadRequestNotFound
		}
		return err
	}

	// Check if the user is authorized to upload the chunk
	if tempTransferData.OwnerID != ownerID {
		return customerrors.ErrUnauthorized
	}
	chunkPath := filepath.Join(constants.ChunkDir, transferID.String())
	err = s.filestorage.DeleteAll(c, chunkPath)
	if err != nil {
		return fmt.Errorf("cancel transfer service:failed to delete temp chunk files by transfer id %s: %w", transferID, err)
	}
	err = s.repo.DeleteTempTransferByID(c, transferID)
	if err != nil {
		return err
	}
	return nil

}

func (s *Service) GetAllUploadedChunksIndex(c context.Context, transferID uuid.UUID, ownerID uuid.UUID) ([]int, error) {
	tempTransferData, err := s.repo.FindTempTransferByID(c, transferID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []int{}, customerrors.ErrUploadRequestNotFound
		}
		return []int{}, err
	}

	// Check if the user is authorized to upload the chunk
	if tempTransferData.OwnerID != ownerID {
		return []int{}, customerrors.ErrUnauthorized
	}

	chunkLst, err := s.repo.GetAllUploadedChunksIndex(c, transferID)
	if err != nil {
		return []int{}, err
	}
	return chunkLst, nil

}

func (s *Service) UploadChunkService(c *gin.Context, chunkUploadRequest dto.ChunkUploadDTO) error {
	// Retrieve the temporary file transfer metadata using the ID
	tempTransferData, err := s.repo.FindTempTransferByID(c, chunkUploadRequest.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return customerrors.ErrUploadRequestNotFound
		}
		return err
	}

	// Check if the user is authorized to upload the chunk
	if tempTransferData.OwnerID != chunkUploadRequest.OwnerID {
		return customerrors.ErrUnauthorized
	}

	// Define the file path where the chunk will be saved
	chunkPath := filepath.Join(constants.ChunkDir, chunkUploadRequest.ID.String())
	err = s.filestorage.CreateFolder(c, chunkPath)
	if err != nil {
		return fmt.Errorf("upload chunk service:failed to create chunk folder for tranferID-%s: %w", chunkUploadRequest.ID, err)

	}
	chunkFilePath := filepath.Join(chunkPath, strconv.Itoa(chunkUploadRequest.ChunkIndex))
	chunkfile, err := chunkUploadRequest.FileChunk.Open()
	if err != nil {
		return fmt.Errorf("upload chunk service:failed to open chunk file in request: index-%d tranferID-%s: %w", chunkUploadRequest.ChunkIndex, chunkUploadRequest.ID, err)
	}
	defer chunkfile.Close()

	// Save the chunk to disk
	chunkwriter, err := s.filestorage.WriteFile(c, chunkFilePath)
	if err != nil {
		return fmt.Errorf("upload chunk service:failed to open chunk file in storage in write mode: index-%d tranferID-%s: %w", chunkUploadRequest.ChunkIndex, chunkUploadRequest.ID, err)
	}
	_, err = io.Copy(chunkwriter, chunkfile)
	if err != nil {
		return fmt.Errorf("upload chunk service:failed to copy data from request chunk file to storage chunk file: index-%d tranferID-%s: %w", chunkUploadRequest.ChunkIndex, chunkUploadRequest.ID, err)
	}
	defer chunkwriter.Close()

	chunk := models.Chunk{
		Index:     chunkUploadRequest.ChunkIndex,
		TranferID: chunkUploadRequest.ID,
	}
	err = s.repo.CreateChunk(c, chunk)
	if err != nil {
		return err
	}

	// Update the last modified time of the temporary transfer record
	err = s.repo.UpdateTempTransferLastUpdatedTimeByID(c, chunkUploadRequest.ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) AssembleFileService(c context.Context, assembleRequest dto.FileAssembleDTO) (uuid.UUID, error) {
	// Fetch temp file metadata
	tempTransferData, err := s.repo.FindTempTransferByID(c, assembleRequest.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.UUID{}, customerrors.ErrUploadRequestNotFound
		}
		return uuid.UUID{}, err
	}

	// Authorization check
	if tempTransferData.OwnerID != assembleRequest.OwnerID {
		return uuid.UUID{}, customerrors.ErrUnauthorized
	}

	chunkPath := filepath.Join(constants.ChunkDir, assembleRequest.ID.String())
	err = s.filestorage.CreateFolder(c, chunkPath)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("assemble file service:failed to create chunk folder for tranferID-%s: %w", assembleRequest.ID, err)

	}

	TempPath := filepath.Join(constants.TempDir, assembleRequest.ID.String())

	err = s.filestorage.CreateFolder(c, TempPath)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("assemble file service:failed to create temp folder for tranferID-%s: %w", assembleRequest.ID, err)

	}

	transferPath := filepath.Join(constants.UploadDir, assembleRequest.ID.String())
	err = s.filestorage.CreateFolder(c, transferPath)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("assemble file service:failed to create upload folder for tranferID-%s: %w", assembleRequest.ID, err)

	}

	finalZipPath, err := s.AssembleAllChunk(c, chunkPath, TempPath)
	if err != nil {
		return uuid.UUID{}, err
	}
	time.Sleep(2 * time.Second)
	// Delete the zip file after extraction
	err = s.filestorage.DeleteAll(c, chunkPath)

	if err != nil {
		return uuid.UUID{}, fmt.Errorf("assemble service:failed to delete temp chunk files: %w", err)
	}
	// Decompress the zip file

	err = utils.Unzip(c, s.filestorage, finalZipPath, transferPath)
	if err != nil {
		return uuid.UUID{}, err
	}
	err = s.filestorage.DeleteAll(c, finalZipPath)
	if err != nil {
		return uuid.UUID{}, err
	}

	// Create and store final transfer record

	expiryTime, err := utils.ParseExpiry(&tempTransferData.Expiry)
	if err != nil {
		return uuid.UUID{}, err
	}

	transferData := models.Transfer{
		ID:           tempTransferData.ID,
		Message:      tempTransferData.Message,
		Expiry:       expiryTime,
		TransferPath: transferPath,
		OwnerID:      tempTransferData.OwnerID,
		Size:         tempTransferData.Size,
	}

	transferID, err := s.repo.CreateTransfer(c, transferData)
	if err != nil {
		return uuid.UUID{}, err
	}

	filesInTransferPath, err := s.filestorage.ReadFolder(c, transferPath)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("assemble service:failed to read extracted files: %w", err)
	}

	for _, f := range filesInTransferPath {
		var fileData models.File
		if !f.IsDir {
			fileData.FileName = f.Name
			fileData.FilePath = filepath.Join(transferPath, f.Name)
			fileData.TransferID = transferID
			fileData.FileSize = f.Size
			fileData.FileExtension = filepath.Ext(f.Name)
			_, err = s.repo.CreateFile(c, fileData)
			if err != nil {
				return uuid.UUID{}, err
			}

		}
	}
	err = s.repo.DeleteTempTransferByID(c, tempTransferData.ID)
	if err != nil {
		return uuid.UUID{}, err
	}
	return transferID, nil
}

func (s *Service) AssembleAllChunk(c context.Context, chunkPath string, tempPath string) (string, error) {
	// Read chunk files
	files, err := s.filestorage.ReadFolder(c, chunkPath)

	if err != nil {
		return "", fmt.Errorf("assemble chunk service:failed to read chunk directory: %w", err)
	}
	err = s.filestorage.CreateFolder(c, tempPath)
	if err != nil {
		return "", fmt.Errorf("assemble chunk service:failed to create upload directory: %w", err)
	}

	// Sort chunk files numerically by filename
	chunkFiles := make([]string, 0)
	for _, file := range files {
		if !file.IsDir {
			chunkFiles = append(chunkFiles, file.Name)
		}
	}
	sort.Slice(chunkFiles, func(i, j int) bool {
		iInt, _ := strconv.Atoi(chunkFiles[i])
		jInt, _ := strconv.Atoi(chunkFiles[j])
		return iInt < jInt
	})

	// Create the final zip file from chunks

	err = s.filestorage.CreateFolder(c, tempPath)
	if err != nil {
		return "", fmt.Errorf("assemble chunk service:failed to create transfer folder: %w", err)
	}

	finalZipPath := filepath.Join(tempPath, "temp.zip")
	outFile, err := s.filestorage.WriteFile(c, finalZipPath)
	if err != nil {
		return "", fmt.Errorf("assemble chunk service:error in opening final zip file for write: %w", err)
	}
	defer func() {
		outFile.Close()
	}()

	// Merge chunks into final zip
	for _, chunk := range chunkFiles {
		chunkFilePath := filepath.Join(chunkPath, chunk)
		chunkData, err := s.filestorage.ReadFile(c, chunkFilePath)
		if err != nil {
			return "", fmt.Errorf("assemble chunk service:failed to open chunk %s: %w", chunk, err)
		}
		_, err = io.Copy(outFile, chunkData)
		if err != nil {
			return "", fmt.Errorf("assemble chunk service:failed to write chunk %s to output file %s : %w", chunk, finalZipPath, err)
		}
		err = chunkData.Close() // close immediately
		fmt.Println("chunk closed")
		if err != nil {
			return "", fmt.Errorf("assemble chunk service: failed to close chunk %s: %w", chunk, err)
		}

	}

	return finalZipPath, nil

}

