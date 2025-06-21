package v1

import (
	"errors"
	"fmt"
	"io"
	"large_fss/internals/constants"
	customerrors "large_fss/internals/customErrors"
	"large_fss/internals/dto"
	"large_fss/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)
func (h *Handler) CreateTransferHandler(c *gin.Context) {
	userIDStr, userExists := c.Get(constants.ClaimPrimaryKey)
	if !userExists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"message": customerrors.ErrUnauthorized.Error()},
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": customerrors.ErrInvalidId.Error()},
		})
		return
	}

	var uploadDTO dto.TransferDTO
	if err := c.BindJSON(&uploadDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": customerrors.ErrInvalidInput.Error()},
		})
		return
	}

	if uploadDTO.Size > int64(constants.ValidUserMaxUploadSize) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": customerrors.LimitExceeded.Error() +
					" - Max Upload Limit: " + strconv.Itoa(constants.ValidUserMaxUploadSize),
			},
		})
		return
	}
	fmt.Println(uploadDTO, "-upload dto")

	if _, err := utils.ParseExpiry(&uploadDTO.Expiry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": customerrors.ErrInvalidInput.Error()},
		})
		return
	}

	uploadDTO.OwnerID = userID

	fileID, err := h.ser.CreateTransferService(c, uploadDTO)
	if err != nil {
		utils.LogErrorWithStack(c, "Internal Server Error (Error in Uploading)", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": customerrors.ErrInternalServer.Error()},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":        constants.SuccessMessage,
		"transfer_id":    fileID,
		"max_chunk_size": constants.MaxChunkSize,
	})
}

func (h *Handler) GetAllUploadedChunksIndexHandler(c *gin.Context) {
	userIDStr, userExists := c.Get(constants.ClaimPrimaryKey)
	if !userExists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"message": customerrors.ErrUnauthorized.Error()},
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": customerrors.ErrInvalidId.Error()},
		})
		return
	}

	transferIDstr := c.Param("transferid")
	transferID, err := uuid.Parse(transferIDstr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": customerrors.ErrInvalidId.Error()},
		})
		return
	}
	chunkLst, err := h.ser.GetAllUploadedChunksIndex(c, transferID, userID)
	if err != nil {
		if errors.Is(err, customerrors.ErrUploadRequestNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"message": customerrors.ErrUploadRequestNotFound.Error()},
			})
			return

		} else if errors.Is(err, customerrors.ErrUnauthorized) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"message": customerrors.ErrUnauthorized.Error()},
			})
			return

		} else {
			utils.LogErrorWithStack(c, "Internal Server Error (Error in Getting all success chunks)", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{"message": customerrors.ErrInternalServer.Error()},
			})
			return

		}

	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":   constants.SuccessMessage,
		"chunk_lst": chunkLst,
	})

}
func (h *Handler) CancelTransferHandler(c *gin.Context) {
	userIDStr, userExists := c.Get(constants.ClaimPrimaryKey)
	if !userExists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"message": customerrors.ErrUnauthorized.Error()},
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": customerrors.ErrInvalidId.Error()},
		})
		return
	}

	var cancelDTO dto.CancelTransferDTO
	if err := c.BindJSON(&cancelDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": customerrors.ErrInvalidInput.Error()},
		})
		return
	}
	transferID, err := uuid.Parse(cancelDTO.ID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": customerrors.ErrInvalidId.Error()},
		})
		return
	}

	err = h.ser.CancelTransferService(c, transferID, userID)
	if err != nil {
		utils.LogErrorWithStack(c, "Internal Server Error (Error in Cancelling)", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": customerrors.ErrInternalServer.Error()},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": constants.SuccessMessage,
	})
}

func (h *Handler) UploadChunkHandler(c *gin.Context) {
	userIDStr, userExists := c.Get(constants.ClaimPrimaryKey)
	if !userExists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"message": customerrors.ErrUnauthorized.Error()},
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": customerrors.ErrInvalidId.Error()},
		})
		return
	}

	uploadIDStr := c.PostForm("uploadId")
	uploadID, err := uuid.Parse(uploadIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": customerrors.ErrInvalidId.Error()},
		})
		return
	}

	chunkIndexStr := c.PostForm("index")
	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": customerrors.ErrInvalidInput.Error()},
		})
		return
	}

	file, err := c.FormFile("chunk")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": customerrors.ErrInvalidInput.Error()},
		})
		return
	}

	uploadChunkDTO := dto.ChunkUploadDTO{
		ChunkIndex: chunkIndex,
		FileChunk:  file,
		ID:         uploadID,
		OwnerID:    userID,
	}

	if err := h.ser.UploadChunkService(c, uploadChunkDTO); err != nil {
		switch {
		case errors.Is(err, customerrors.ErrUnauthorized):
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"message": customerrors.ErrUnauthorized.Error()},
			})
		case errors.Is(err, customerrors.ErrUploadRequestNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"message": customerrors.ErrUploadRequestNotFound.Error()},
			})
		default:
			utils.LogErrorWithStack(c, "Internal Server Error (Error in Chunk Upload)", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{"message": customerrors.ErrInternalServer.Error()},
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": constants.SuccessMessage,
	})
}

func (h *Handler) AssembleFileHandler(c *gin.Context) {
	userIDStr, userExists := c.Get(constants.ClaimPrimaryKey)
	if !userExists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"message": customerrors.ErrUnauthorized.Error()},
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": customerrors.ErrInvalidId.Error()},
		})
		return
	}

	var assembleDTO dto.FileAssembleDTO
	if err := c.BindJSON(&assembleDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": customerrors.ErrInvalidInput.Error()},
		})
		return
	}

	assembleDTO.OwnerID = userID

	fileID, err := h.ser.AssembleFileService(c, assembleDTO)
	if err != nil {
		utils.LogErrorWithStack(c, "Internal Server Error in AssembleFileService", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": customerrors.ErrInternalServer.Error()},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     constants.SuccessMessage,
		"transfer_id": fileID,
	})
}
func (h *Handler) GetTransferInfoHandler(c *gin.Context) {
	transferIDstr := c.Param("transferid")
	transferID, err := uuid.Parse(transferIDstr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": customerrors.ErrInvalidId.Error()},
		})
		return
	}
	transferinfo, err := h.ser.GetTransferInfoService(c, transferID)
	if err != nil {
		if errors.Is(err, customerrors.ErrExpiredLink) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"message": customerrors.ErrExpiredLink.Error()},
			})
			return

		} else if errors.Is(err, customerrors.ErrFileNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"message": customerrors.ErrFileNotFound.Error()},
			})
			return

		} else {
			utils.LogErrorWithStack(c, "Internal Server Error in AssembleFileService", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{"message": customerrors.ErrInternalServer.Error()},
			})
			return

		}
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": constants.SuccessMessage,
		"data":    transferinfo,
	})

}

func (h *Handler) FileDownloaderHandler(c *gin.Context) {
	fileIDstr := c.Param("fileid")
	fileID, err := uuid.Parse(fileIDstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": customerrors.ErrInvalidInput.Error()},
		})
		return
	}

	// Get the file path and deletion flag from service
	file, filename, err := h.ser.FileDownloaderService(c, fileID)
	if err != nil {
		if errors.Is(err, customerrors.ErrFileNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"message": customerrors.ErrFileNotFound.Error()},
			})
			return

		} else {
			utils.LogErrorWithStack(c, "Internal Server Error in Getting file path for transfer downloader", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{"message": customerrors.ErrInternalServer.Error()},
			})
			return

		}

	}

	// Ensure file is closed and optionally deleted after response is sent
	defer func() {
		file.Close()
	}()

	// Set headers for file download
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", "application/octet-stream") // Or use http.DetectContentType for dynamic type

	// Stream file to client
	_, err = io.Copy(c.Writer, file)
	if err != nil {
		utils.LogErrorWithStack(c, "Internal Server Error in Streaming Content", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": customerrors.ErrInternalServer.Error()},
		})
		return
	}

}

func (h *Handler) TransferDownloaderHandler(c *gin.Context) {
	transferIDstr := c.Param("transferid")
	transferID, err := uuid.Parse(transferIDstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": customerrors.ErrInvalidId.Error()},
		})
		return
	}

	// Get the file path and deletion flag from service
	file, filename, err := h.ser.TransferDownloaderService(c, transferID)
	if err != nil {

		if errors.Is(err, customerrors.ErrExpiredLink) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"message": customerrors.ErrExpiredLink.Error()},
			})
			return

		} else if errors.Is(err, customerrors.ErrFileNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"message": customerrors.ErrFileNotFound.Error()},
			})
			return

		} else {
			utils.LogErrorWithStack(c, "Internal Server Error in Getting file path for transfer downloader", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{"message": customerrors.ErrInternalServer.Error()},
			})
			return

		}

	}

	// Ensure file is closed and optionally deleted after response is sent
	defer func() {
		file.Close()
	}()

	// Set headers for file download
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", "application/octet-stream") // Or use http.DetectContentType for dynamic type

	// Stream file to client
	_, err = io.Copy(c.Writer, file)
	if err != nil {
		utils.LogErrorWithStack(c, "Internal Server Error in Streaming Content", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": customerrors.ErrInternalServer.Error()},
		})
		return
	}
}

func (h *Handler) GetAllTransfersHandler(c *gin.Context) {
	userIDStr, userExists := c.Get(constants.ClaimPrimaryKey)
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": customerrors.ErrInvalidId.Error(),
			},
		})
		return
	}

	if !userExists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"message": customerrors.ErrUnauthorized.Error(),
			},
		})
		return
	}
	transferlst, err := h.ser.GetAllTransfersService(c, userID)
	if err != nil {
		utils.LogErrorWithStack(c, "Internal Server Error in AssembleFileService", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": customerrors.ErrInternalServer.Error()},
		})
		return

	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": constants.SuccessMessage,
		"data":    transferlst,
	})

}

func (h *Handler) UpdateTransferHandler(c *gin.Context) {
	userIDStr, userExists := c.Get(constants.ClaimPrimaryKey)
	if !userExists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"message": customerrors.ErrUnauthorized.Error()},
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": customerrors.ErrInvalidId.Error()},
		})
		return
	}

	var updateDTO dto.TransferUpdateDTO
	if err := c.BindJSON(&updateDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": customerrors.ErrInvalidInput.Error()},
		})
		return
	}

	updateDTO.OwnerID = userID
	err = h.ser.UpdateTransferService(c, updateDTO)
	if err != nil {
		if errors.Is(err, customerrors.ErrExpiredLink) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"message": customerrors.ErrExpiredLink.Error()},
			})
			return

		} else if errors.Is(err, customerrors.ErrUnauthorized) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"message": customerrors.ErrUnauthorized.Error()},
			})
			return

		} else if errors.Is(err, customerrors.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{"message": customerrors.ErrInvalidInput.Error()},
			})
			return

		} else {
			utils.LogErrorWithStack(c, "Internal Server Error in Getting file path for transfer downloader", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{"message": customerrors.ErrInternalServer.Error()},
			})
			return

		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": constants.SuccessMessage,
	})

}

func (h *Handler)DeleteTransferHandler(c *gin.Context){
	userIDStr, userExists := c.Get(constants.ClaimPrimaryKey)
	if !userExists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"message": customerrors.ErrUnauthorized.Error()},
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": customerrors.ErrInvalidId.Error()},
		})
		return
	}
	transferIDstr := c.Param("transferid")
	transferID, err := uuid.Parse(transferIDstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": customerrors.ErrInvalidId.Error()},
		})
		return
	}
	err=h.ser.DeleteTransferService(c,transferID,userID)
	if(err!=nil){
			if errors.Is(err, customerrors.ErrExpiredLink) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"message": customerrors.ErrExpiredLink.Error()},
			})
			return

		} else if errors.Is(err, customerrors.ErrUnauthorized) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"message": customerrors.ErrUnauthorized.Error()},
			})
			return

		}  else {
			utils.LogErrorWithStack(c, "Internal Server Error in Getting file path for transfer downloader", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{"message": customerrors.ErrInternalServer.Error()},
			})
			return

		}

	}
		c.JSON(http.StatusAccepted, gin.H{
		"message": constants.SuccessMessage,
	})

}