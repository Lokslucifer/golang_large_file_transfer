package customerrors

import (
	"errors"
)

var (
	ErrFileUpload     = errors.New("error in file uploading")
	ErrInfoNotFound   = errors.New("error info not found in databas")
	ErrInvalidPdfFile = errors.New("invalid file format(should be in pdf)")
	ErrFileNotFound   = errors.New("file id invalid or file removed")
	ErrUploadRequestNotFound=errors.New("upload request is canceled or removed")
	ErrInvalidFilter=errors.New("invalid filter")
	LimitExceeded=errors.New("limit Exceed")
	ErrInvalidLink=errors.New("invalid link")
	ErrExpiredLink=errors.New("link is expired or deleted")

)
