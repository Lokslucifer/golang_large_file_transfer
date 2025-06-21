package dto

type SignUpDTO struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type LoginDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}


