package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"large_fss/internals/constants"
	customerrors "large_fss/internals/customErrors"
	"large_fss/internals/dto"
	"large_fss/utils"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *Service) GetTransferInfoService(c context.Context, transferID uuid.UUID) (*dto.TransferInfoDTO, error) {
	transferData, err := s.repo.FindTransferByID(c, transferID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customerrors.ErrExpiredLink
		}
		return nil, err
	}
	filesData, err := s.repo.FindAllFilesByTransferID(c, transferID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customerrors.ErrFileNotFound
		}
		return nil, err
	}

	var fileInfoList []dto.FileInfoDTO
	for _, file := range filesData {
		fileinfo := dto.FileInfoDTO{
			ID:            file.ID,
			FileName:      file.FileName,
			FileSize:      file.FileSize,
			FileExtension: file.FileExtension,
		}
		fileInfoList = append(fileInfoList, fileinfo)
	}
	transferInfo := dto.TransferInfoDTO{
		ID:           transferData.ID,
		Message:      transferData.Message,
		Size:         int64(transferData.Size),
		Expiry:       *transferData.Expiry,
		FileInfoList: fileInfoList,
	}
	return &transferInfo, nil

}

func (s *Service) TransferDownloaderService(c *gin.Context, transferID uuid.UUID) (io.ReadCloser, string, error) {
	// Retrieve transfer metadata
	transferData, err := s.repo.FindTransferByID(c, transferID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", customerrors.ErrExpiredLink
		}
		return nil, "", err
	}

	// Retrieve all files associated with the transfer
	filesData, err := s.repo.FindAllFilesByTransferID(c, transferID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", customerrors.ErrFileNotFound
		}
		return nil, "", err
	}

	// Single file: stream directly
	if len(filesData) == 1 {
		resourcePath := filesData[0].FilePath
		reader, err := s.filestorage.ReadFile(c, resourcePath)
		_, filename := filepath.Split(resourcePath)
		if err != nil {
			return nil, "", err
		}
		err = s.repo.IncrementActiveStreamByID(c, filesData[0].ID)
		if err != nil {
			return nil, "", fmt.Errorf("transfer downloader service:failed to increment active stream for fileID-%s: %w", filesData[0].ID, err)

		}
		wrappedfileReader := &autoFileReader{
			ReadCloser: reader,
			repo:       s.repo,
			FileID:     filesData[0].ID,
			ctx:        c,
		}

		return wrappedfileReader, filename, nil
	}
	err = s.filestorage.CreateFolder(c, constants.TempDir)
	if err != nil {
		return nil, "", fmt.Errorf("transfer downloader service:failed to create transfer folder for tranferID-%s: %w", transferID, err)
	}

	// Multiple files: zip them, then stream
	tempTransferZipPath := filepath.Join(constants.TempDir, transferID.String()+".zip")
	transferPath := transferData.TransferPath

	fmt.Println("transfer path -", transferPath)

	err = utils.CreateZip(c, s.filestorage, transferPath, tempTransferZipPath)
	if err != nil {
		return nil, "", err
	}

	reader, err := s.filestorage.ReadFile(c, tempTransferZipPath)
	if err != nil {
		return nil, "", err
	}

	wrappedReader := &autoDeleteReader{
		ReadCloser:  reader,
		path:        tempTransferZipPath,
		fileStorage: s.filestorage,
		ctx:         c,
	}
	_, filename := filepath.Split(tempTransferZipPath)

	return wrappedReader, filename, nil
}

func (s *Service) FileDownloaderService(c *gin.Context, fileID uuid.UUID) (io.ReadCloser, string, error) {
	fileData, err := s.repo.FindFileByID(c, fileID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", customerrors.ErrFileNotFound
		}
		return nil, "", err
	}
	reader, err := s.filestorage.ReadFile(c, fileData.FilePath)
	if err != nil {
		return nil, "", err
	}
	_, filename := filepath.Split(fileData.FilePath)
	err = s.repo.IncrementActiveStreamByID(c, fileData.ID)
	if err != nil {
		return nil, "", fmt.Errorf("transfer downloader service:failed to increment active stream for fileID-%s: %w", fileData.ID, err)

	}
	wrappedReader := &autoFileReader{
		ReadCloser: reader,
		repo:       s.repo,
		FileID:     fileData.ID,
		ctx:        c,
	}

	return wrappedReader, filename, nil
}

func (s *Service) GetAllTransfersService(c context.Context, userID uuid.UUID) ([]dto.TransferInfoDTO, error) {
	transferLst, err := s.repo.FindAllTransfersByUserID(c, userID)
	if err != nil {
		return []dto.TransferInfoDTO{}, err

	}
	var transferDTOLst []dto.TransferInfoDTO

	for _, trans := range transferLst {
		transDTO := dto.TransferInfoDTO{
			ID:        trans.ID,
			Expiry:    *trans.Expiry,
			Message:   trans.Message,
			Size:      trans.Size,
			CreatedAt: trans.CreatedAt,
		}
		transferDTOLst = append(transferDTOLst, transDTO)
	}
	return transferDTOLst, nil

}

func (s *Service) UpdateTransferService(c context.Context, updateDTO dto.TransferUpdateDTO) error {
	transferData, err := s.repo.FindTransferByID(c, updateDTO.TransferID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return customerrors.ErrExpiredLink
		}
		return err
	}
	if transferData.OwnerID != updateDTO.OwnerID {
		return customerrors.ErrUnauthorized

	}
	transferData.Message = updateDTO.Message
	if updateDTO.Expiry != "" {
		expiryTime, err := utils.ParseExpiry(&updateDTO.Expiry)
		if err != nil {
			return customerrors.ErrInvalidInput
		}
		transferData.Expiry = expiryTime
	}
	err = s.repo.UpdateTransferByID(c, *transferData)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) DeleteTransferService(c context.Context, transferID uuid.UUID, userID uuid.UUID) error {
	transferData, err := s.repo.FindTransferByID(c, transferID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return customerrors.ErrExpiredLink
		}
		return err
	}
	if transferData.OwnerID != userID {
		return customerrors.ErrUnauthorized

	}
	err = s.filestorage.DeleteAll(c, transferData.TransferPath)
	if err != nil {
		return fmt.Errorf("delete transfer service: failed to remove/delete path %s: %w", transferData.TransferPath, err)
	}
	err = s.repo.DeleteTransferByID(c, transferID)
	if err != nil {
		return err
	}
	return nil
}
