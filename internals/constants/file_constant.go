package constants

//Resource Constants
const (
	UploadDir     = "uploads"     // Directory to store final assembled files
	ChunkDir      = "chunks"      // Directory to store individual chunks
	TempDir	  ="temp"
	MaxChunkSize     = 5*1024 * 1024   // 1MB chunk size (example, can be adjusted)
	ValidUserMaxUploadSize = 5 * 1024 * 1024 * 1024 // 5GB max file size (example)
	NonUserMaxUploadSize=1*1024*1024*1024
	MaxhoursUploadSessionValid=4
	//error messages
	ErrInvalidFileFormat = "Invalid file format"


)
