package utils

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"large_fss/internals/storage"
	"net/http"
	"path"
	"path/filepath"
	"time"
)

func ParseExpiry(expiryStr *string) (*time.Time, error) {
	if *expiryStr == "never" {
		// Represents "never expires"
		return nil, nil
	}

	now := time.Now().UTC()
	var expiry time.Time

	switch *expiryStr {
	case "5m":
		expiry = now.Add(5 * time.Minute)
	case "3h":
		expiry = now.Add(3 * time.Hour)
	case "12h":
		expiry = now.Add(12 * time.Hour)
	case "1d":
		expiry = now.Add(24 * time.Hour)
	case "3d":
		expiry = now.Add(72 * time.Hour)
	case "1w":
		expiry = now.Add(7 * 24 * time.Hour)
	default:
		return nil, errors.New("parse expiry:invalid expiry value")
	}

	return &expiry, nil
}

func Unzip(ctx context.Context, storage storage.Storage, srcPath string, destPath string) error {
	// Open the zip file using the storage interface
	srcReader, err := storage.ReadFile(ctx, srcPath)
	if err != nil {
		return fmt.Errorf("unzip util:failed to open zip file from storage: %w", err)
	}
	defer srcReader.Close()

	// Read all bytes of the zip file
	zipBytes, err := io.ReadAll(srcReader)
	if err != nil {
		return fmt.Errorf("unzip util:failed to read zip file: %w", err)
	}

	// Create a zip reader from the bytes
	zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return fmt.Errorf("unzip util:failed to create zip reader: %w", err)
	}

	// Iterate over files in zip
	for _, f := range zipReader.File {
		fPath := path.Join(destPath, f.Name)
		

		if f.FileInfo().IsDir() {
			err := storage.CreateFolder(ctx, fPath)
			fmt.Println("unzip creating folder -",f)
			if err != nil {
				return fmt.Errorf("unzip util:failed to create folder %s: %w", fPath, err)
			}
			continue
		}

		// Open the file inside the zip
		fileInZip, err := f.Open()
		if err != nil {
			return fmt.Errorf("unzip util:failed to open file %s in zip: %w", f.Name, err)
		}

		// Create destination file in storage
		writer, err := storage.WriteFile(ctx, fPath)
		if err != nil {
			fileInZip.Close()
			return fmt.Errorf("unzip util:failed to create file %s in storage: %w", fPath, err)
		}

		// Copy the content
		_, err = io.Copy(writer, fileInZip)
		fileInZip.Close()
		writer.Close()

		if err != nil {
			return fmt.Errorf("unzip util:failed to copy contents to %s: %w", fPath, err)
		}
	}

	return nil
}

// CreateZipFromFolder takes a folder path, compresses it into a zip, and returns the zip file path.
func CreateZip(ctx context.Context, storage storage.Storage, folderPath string, outputZipPath string) error {
	// Ensure output zip file is created
	err := storage.CreateFile(ctx, outputZipPath)
	if err != nil {
		return fmt.Errorf("create zip util:failed to create zip file: %w", err)
	}

	// Open the zip file for writing
	zipWriterCloser, err := storage.WriteFile(ctx, outputZipPath)
	if err != nil {
		return fmt.Errorf("create zip util:failed to open zip file for writing: %w", err)
	}
	defer zipWriterCloser.Close()

	zipWriter := zip.NewWriter(zipWriterCloser)
	defer zipWriter.Close()

	// Get list of all files and directories inside the folder
	files, err := storage.ListFilesRecursive(ctx, folderPath)
	if err != nil {
		return fmt.Errorf("create zip util:failed to list folder contents: %w", err)
	}

	for _, file := range files {
		fmt.Println(folderPath,"-",file.Path)
		relPath, err := filepath.Rel(folderPath, file.Path)
		fmt.Println("rel path-",relPath)
		if err != nil {
			return fmt.Errorf("create zip util:failed to compute relative path: %w", err)
		}

		if file.IsDir {
			// Add directory with trailing slash
			_, err := zipWriter.Create(relPath + "/")
			if err != nil {
				return fmt.Errorf("create zip util:failed to add directory to zip: %w", err)
			}
			continue
		}

		// Read file from storage
		reader, err := storage.ReadFile(ctx, file.Path)
		if err != nil {
			return fmt.Errorf("create zip util:failed to read file %s: %w", file.Path, err)
		}

		writer, err := zipWriter.Create(relPath)
		if err != nil {
			reader.Close()
			return fmt.Errorf("create zip util:failed to create zip entry for %s: %w", relPath, err)
		}

		_, err = io.Copy(writer, reader)
		reader.Close()
		if err != nil {
			return fmt.Errorf("create zip util:failed to write file %s to zip: %w", relPath, err)
		}
	}

	return nil
}

func DetectContentTypeFromReader(r io.ReadCloser) (string, error) {

	buffer := &bytes.Buffer{}
	tee := io.TeeReader(io.LimitReader(r, 512), buffer)

	// Read 512 bytes
	buf := make([]byte, 512)
	_, err := io.ReadFull(tee, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "",fmt.Errorf("detect content type util:error in reading content %w", err)
	}

	contentType := http.DetectContentType(buf)

	// Optionally return a new reader with full contents if needed
	return contentType, nil
}
