package storage

import (
	"context"
	"io"
	"large_fss/internals/models"
	
)

type Storage interface {
	CreateFile(ctx context.Context, filePath string) error
	ReadFile(ctx context.Context, filePath string) (io.ReadCloser, error)
	WriteFile(ctx context.Context, filePath string) (io.WriteCloser, error)

	CreateFolder(ctx context.Context, folderPath string) error
	ReadFolder(ctx context.Context, folderPath string) ([]models.SysFileInfo, error)
	IsFolder(ctx context.Context, folderPath string) (bool, error)
	ListFilesRecursive(ctx context.Context, folderPath string) ([]models.SysFileInfo, error)

	DeleteAll(ctx context.Context, path string) error
	DeleteFile(ctx context.Context, filePath string) error     // Optional
	DeleteFolder(ctx context.Context, folderPath string) error // Optional

	Exists(ctx context.Context, path string) (bool, error)      // Optional
	Stat(ctx context.Context, path string) (models.SysFileInfo, error) // Optional

}
