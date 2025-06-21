package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"large_fss/internals/models"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	BaseDir string
}

func NewLocalStorage(baseDir string) Storage {
	return &LocalStorage{BaseDir: baseDir}
}

// Helper to resolve relative paths to full paths under BaseDir
func (l *LocalStorage) resolve(path string) string {
	return filepath.Join(l.BaseDir, path)
}

// CreateFile creates an empty file at the specified path.
func (l *LocalStorage) CreateFile(ctx context.Context, filePath string) error {
	fullPath := l.resolve(filePath)
	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	return f.Close()
}

// ReadFile opens a file for reading.
func (l *LocalStorage) ReadFile(ctx context.Context, filePath string) (io.ReadCloser, error) {
	fmt.Println("reading-", filePath)
	return os.Open(l.resolve(filePath))
}

// WriteFile opens a file for writing.
func (l *LocalStorage) WriteFile(ctx context.Context, filePath string) (io.WriteCloser, error) {
	fmt.Println("reading-", filePath)
	return os.Create(l.resolve(filePath))
}

// CreateFolder creates a new directory and all necessary parents.
func (l *LocalStorage) CreateFolder(ctx context.Context, folderPath string) error {
	return os.MkdirAll(l.resolve(folderPath), os.ModePerm)
}

// ReadFolder returns the list of files and folders in a directory.
func (l *LocalStorage) ReadFolder(ctx context.Context, folderPath string) ([]models.SysFileInfo, error) {
	entries, err := os.ReadDir(l.resolve(folderPath))
	if err != nil {
		return nil, err
	}

	var fileInfos []models.SysFileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}

		fileInfos = append(fileInfos, models.SysFileInfo{
			Name:    info.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   info.IsDir(),
		})
	}

	return fileInfos, nil
}

// ListFilesRecursive recursively lists all files and folders under a directory.
func (l *LocalStorage) ListFilesRecursive(ctx context.Context, folderPath string) ([]models.SysFileInfo, error) {
	var files []models.SysFileInfo
	fullPath := l.resolve(folderPath)

	err := filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == fullPath {
			return nil
		}

		relPath, _ := filepath.Rel(l.BaseDir, path)

		files = append(files, models.SysFileInfo{
			Name:    info.Name(),
			Path:    relPath,
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   info.IsDir(),
		})
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// IsFolder checks whether a path is a directory.
func (l *LocalStorage) IsFolder(ctx context.Context, folderPath string) (bool, error) {
	info, err := os.Stat(l.resolve(folderPath))
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// DeleteAll deletes a file or folder and its contents.
func (l *LocalStorage) DeleteAll(ctx context.Context, path string) error {
	return os.RemoveAll(l.resolve(path))
}

// DeleteFile deletes a single file.
func (l *LocalStorage) DeleteFile(ctx context.Context, filePath string) error {
	return os.Remove(l.resolve(filePath))
}

// DeleteFolder deletes a directory.
func (l *LocalStorage) DeleteFolder(ctx context.Context, folderPath string) error {
	return os.RemoveAll(l.resolve(folderPath))
}

// Exists checks whether a file or folder exists.
func (l *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
	_, err := os.Stat(l.resolve(path))
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

// Stat returns file or folder information.
func (l *LocalStorage) Stat(ctx context.Context, path string) (models.SysFileInfo, error) {
	info, err := os.Stat(l.resolve(path))
	if err != nil {
		return models.SysFileInfo{}, err
	}
	return models.SysFileInfo{
		Name:  info.Name(),
		Size:  info.Size(),
		IsDir: info.IsDir(),
	}, nil
}
