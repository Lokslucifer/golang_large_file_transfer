package services

import (
	"fmt"
	"large_fss/internals/constants"
	customerrors "large_fss/internals/customErrors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type JWTService struct {
	secret string
}

func NewJWTService() (*JWTService, error) {
	secret := os.Getenv("JWT_SECRET")
	fmt.Println(secret, "-secret key")

	if secret == "" {
		return nil, customerrors.ErrSecretKeyNotFound
	}
	return &JWTService{secret: secret}, nil
}
func (j *JWTService) CreateJWT(id uuid.UUID) (string, error) {
	// Create a new token with signing method and claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		constants.ClaimPrimaryKey: id,
		"nbf":                     time.Now().UTC().Unix(),
	})

	secret := j.secret
	if secret == "" {
		return "", customerrors.ErrSecretKeyNotFound
	}
	tokenstr, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("jwt service: create jwt token: %w", err)
	}
	return tokenstr, nil
}

func (j *JWTService) ValidateJWT(tokenString string) (*jwt.MapClaims, error) {
	fmt.Println("val secret key-", j.secret)
	// Parse the JWT token with validation
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC-SHA256
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("jwt service: unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.secret), nil
	})

	if err != nil {
		log.Printf("JWT parse error: %v", err)
		return nil, customerrors.ErrInvalidToken
	}

	if !token.Valid {
		log.Println("JWT token is invalid")
		return nil, customerrors.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Println("JWT claims conversion failed")
		return nil, customerrors.ErrInvalidToken
	}

	// Optional: Log the primary claim
	if primaryKey, ok := claims[constants.ClaimPrimaryKey]; ok {
		log.Printf("Authenticated user ID: %v", primaryKey)
	}

	return &claims, nil
}
