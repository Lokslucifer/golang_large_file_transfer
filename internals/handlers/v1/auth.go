package v1

import (
	"errors"
	"fmt"
	"large_fss/internals/constants"
	customerrors "large_fss/internals/customErrors"
	"large_fss/internals/dto"
	"large_fss/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) SignupHandler(c *gin.Context) {
	fmt.Println(c.Request.Body)

	var signupRequest dto.SignUpDTO

	err := c.BindJSON(&signupRequest)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": customerrors.ErrBadRequest.Error(),
			},
		})
		return
	}
	fmt.Println(signupRequest)
	userId, err := h.ser.Signup(c, &signupRequest)

	if err != nil {
		if errors.Is(err, customerrors.ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{
				"error": gin.H{
					"message": customerrors.ErrUserAlreadyExists.Error(),
				},
			})
			return
		} else {

			utils.LogErrorWithStack(c, "Internal Server Error in Signup user", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"message": customerrors.ErrInternalServer.Error(),
				},
			})

			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": constants.SuccessMessage,
		"ID":      userId,
	})

}

func (h *Handler) LoginHandler(c *gin.Context) {

	var loginRequest dto.LoginDTO

	err := c.BindJSON(&loginRequest)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": customerrors.ErrInvalidInput.Error(),
			},
		})
		return
	}

	userToken, err := h.ser.Login(c, &loginRequest)
	if err != nil {
		if errors.Is(err, customerrors.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"message": customerrors.ErrUserNotFound.Error(),
				},
			})
			return
		} else if errors.Is(err, customerrors.ErrInvalidPassword) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"message": customerrors.ErrInvalidPassword.Error(),
				},
			})
			return
		} else {
			utils.LogErrorWithStack(c, "Internal Server Error in Login", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"message": customerrors.ErrInternalServer.Error(),
				},
			})
			return
		}
	}
	c.SetCookie(
		"auth_token", // Cookie name
		userToken,    // Value
		3600*6,       // MaxAge in seconds (e.g., 1 hour)
		"/",          // Path
		"",           // Domain ("" uses the request's domain)
		true,         // Secure (set to true in production with HTTPS)
		true,         // HttpOnly (not accessible via JS)
	)

	c.JSON(http.StatusCreated, gin.H{
		"message": constants.SuccessMessage,
		"Token":   userToken,
	})

}
