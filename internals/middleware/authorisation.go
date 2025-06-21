package middlewares

import (
	"large_fss/internals/constants"
	customerrors "large_fss/internals/customErrors"
	"large_fss/internals/services"
	"net/http"

	"github.com/gin-gonic/gin"
)


func  AuthorizationMiddleware(jwtServiceObj *services.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {

		// We will check for the authorization header
		authorization, err := c.Cookie("auth_token")

		if err != nil || authorization == "" {
			// If token not found in cookie, check Authorization header
			authorization = c.GetHeader("auth_token")

			if len(authorization) == 0 {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": gin.H{
						"message": customerrors.ErrMissingToken.Error(),
					},
				})

				c.Redirect(http.StatusFound, "/") // or "/login"
				c.Abort()
				return
			}

		}
		// We will check if the authorization header is valid
		claims, err := jwtServiceObj.ValidateJWT(authorization)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"message": customerrors.ErrInvalidToken.Error(),
				},
			})
			c.Abort()
			return
		}
		c.Set(constants.ClaimPrimaryKey, (*claims)[constants.ClaimPrimaryKey])
		c.Next()
	}
}
