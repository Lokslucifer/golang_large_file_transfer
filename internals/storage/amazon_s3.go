package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"path"
	"strings"

	"large_fss/internals/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Storage struct {
	Client     *s3.Client
	BucketName string
}

func NewS3Storage(client *s3.Client, bucketName string) Storage {
	return &S3Storage{
		Client:     client,
		BucketName: bucketName,
	}
}

func (s *S3Storage) CreateFile(ctx context.Context, filePath string) error {
	_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(filePath),
		Body:   bytes.NewReader([]byte{}), // Empty file
	})
	return err
}

func (s *S3Storage) ReadFile(ctx context.Context, filePath string) (io.ReadCloser, error) {
	
	resp, err := s.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(filePath),
	})
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (s *S3Storage) WriteFile(ctx context.Context, filePath string) (io.WriteCloser, error) {
	pr, pw := io.Pipe()

	go func() {
		_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(s.BucketName),
			Key:    aws.String(filePath),
			Body:   pr,
		})
		_ = pr.CloseWithError(err)
	}()

	return pw, nil
}

func (s *S3Storage) CreateFolder(ctx context.Context, folderPath string) error {
	if !strings.HasSuffix(folderPath, "/") {
		folderPath += "/"
	}
	_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(folderPath),
		Body:   bytes.NewReader([]byte{}),
	})
	return err
}

func (s *S3Storage) ReadFolder(ctx context.Context, folderPath string) ([]models.SysFileInfo, error) {
	if !strings.HasSuffix(folderPath, "/") {
		folderPath += "/"
	}

	resp, err := s.Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.BucketName),
		Prefix: aws.String(folderPath),
	})
	if err != nil {
		return nil, err
	}

	var files []models.SysFileInfo
	for _, obj := range resp.Contents {
		// Skip the folder itself
		if *obj.Key == folderPath {
			continue
		}

		fileName := strings.TrimPrefix(*obj.Key, folderPath)

		files = append(files, models.SysFileInfo{
			Name:    fileName,
			Path:    *obj.Key,
			Size:    *obj.Size,
			ModTime: *obj.LastModified,
			IsDir:   false, // S3 doesn't store directories explicitly
		})
	}

	return files, nil
}

func (s *S3Storage) IsFolder(ctx context.Context, folderPath string) (bool, error) {
	if !strings.HasSuffix(folderPath, "/") {
		folderPath += "/"
	}

	resp, err := s.Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  aws.String(s.BucketName),
		Prefix:  aws.String(folderPath),
		MaxKeys: aws.Int32(1),
	})
	if err != nil {
		return false, err
	}
	return len(resp.Contents) > 0, nil
}

func (s S3Storage) ListFilesRecursive(ctx context.Context, folderPath string) ([]models.SysFileInfo, error) {
	if !strings.HasSuffix(folderPath, "/") {
		folderPath += "/"
	}

	var allFiles []models.SysFileInfo
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.BucketName),
		Prefix: aws.String(folderPath),
	}

	paginator := s3.NewListObjectsV2Paginator(s.Client, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, obj := range page.Contents {
			// Skip the folder itself
			if *obj.Key == folderPath {
				continue
			}

			fileName := path.Base(*obj.Key)
			allFiles = append(allFiles, models.SysFileInfo{
				Name:    fileName,
				Path:    *obj.Key,
				Size:    *obj.Size,
				ModTime: *obj.LastModified,
				IsDir:   false, // S3 has no true directories
			})
		}
	}

	return allFiles, nil
}

func (s *S3Storage) DeleteAll(ctx context.Context, path string) error {
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	resp, err := s.Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.BucketName),
		Prefix: aws.String(path),
	})
	if err != nil {
		return err
	}

	var objects []types.ObjectIdentifier
	for _, item := range resp.Contents {
		objects = append(objects, types.ObjectIdentifier{Key: item.Key})
	}

	if len(objects) == 0 {
		return nil
	}

	_, err = s.Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(s.BucketName),
		Delete: &types.Delete{Objects: objects},
	})
	return err
}

func (s *S3Storage) DeleteFile(ctx context.Context, filePath string) error {
	_, err := s.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(filePath),
	})
	return err
}

func (s *S3Storage) DeleteFolder(ctx context.Context, folderPath string) error {
	return s.DeleteAll(ctx, folderPath)
}

func (s *S3Storage) Exists(ctx context.Context, path string) (bool, error) {
	_, err := s.Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(path),
	})
	if err != nil {
		var nsk *types.NotFound
		if errors.As(err, &nsk) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *S3Storage) Stat(ctx context.Context, path string) (models.SysFileInfo, error) {
	resp, err := s.Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(path),
	})
	if err != nil {
		return models.SysFileInfo{}, err
	}

	return models.SysFileInfo{
		Name:  path,
		Size:  *resp.ContentLength,
		IsDir: strings.HasSuffix(path, "/"),
	}, nil
}
