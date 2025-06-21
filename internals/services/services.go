package services

import (
	"context"
	"fmt"
	"io"
	"large_fss/internals/repository"
	"large_fss/internals/storage"
	"log"

	"github.com/google/uuid"
)

type Service struct {
	JwtService  *JWTService
	repo        repository.DbRepository
	filestorage storage.Storage
}

func NewService(jwtservice *JWTService,repo repository.DbRepository, filestore storage.Storage) *Service {
	return &Service{JwtService: jwtservice, repo: repo, filestorage: filestore}
}

type autoDeleteReader struct {
	io.ReadCloser
	path        string
	fileStorage storage.Storage
	ctx         context.Context
}

type autoFileReader struct {
	io.ReadCloser
	FileID uuid.UUID
	ctx    context.Context
	repo   repository.DbRepository
}

func (r *autoFileReader) Close() error {
	readErr := r.ReadCloser.Close()

	// DecrementActiveStreamByID is called after closing the stream
	streamErr := r.repo.DecrementActiveStreamByID(r.ctx, r.FileID)
	if streamErr != nil {
		// Log stream decrement error, or optionally handle it
		log.Printf("autoFileReader: error decrementing active stream: %v", streamErr)
	}

	// Return the error only if closing ReadCloser failed
	if readErr != nil {
		return fmt.Errorf("autoFileReader: close operation failed: %w", readErr)
	}

	return nil
}

func (r *autoDeleteReader) Close() error {
	readErr := r.ReadCloser.Close()

	// Attempt to delete the file (log or handle the error if needed)
	delErr := r.fileStorage.DeleteFile(r.ctx, r.path)
	if delErr != nil {
		log.Printf("autoDeleteReader: error deleting file at %s: %v", r.path, delErr)
	}

	// Return readErr only if there was an actual error
	if readErr != nil {
		return fmt.Errorf("autoDeleteReader: close operation failed: %w", readErr)
	}

	return nil
}
