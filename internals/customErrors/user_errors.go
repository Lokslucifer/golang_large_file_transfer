package customerrors
import ("errors")

//Server Error Messages
var (
	

	// Error User Messages
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrInvalidPassword = errors.New("invalid password")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrTokenGeneration = errors.New("token generation failed")
	ErrUnauthorized =errors.New("unauthorized") 

	//Error Author Request
	ErrRequestAlreadyExists=errors.New("request already exists")

	//Error Authorization Messages
	ErrInvalidToken = errors.New("invalid token")
	ErrMissingToken =errors.New("missing token") 

	//Error JWT Service
	ErrSecretKeyNotFound = errors.New("secret key not found")
)

