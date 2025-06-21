package repository

import (
	"context"
	"fmt"
	"large_fss/internals/models"

	"github.com/google/uuid"
)

func (r *PostgresSQLDB) CreateUser(ctx context.Context, user models.User) (uuid.UUID, error) {
	query := `INSERT INTO users (email, password, first_name, last_name)
				VALUES ($1, $2, $3, $4)
				RETURNING id` // Use RETURNING to get the new ID

	var userIdStr string
	err := r.db.QueryRowContext(ctx, query, user.Email, user.Password, user.FirstName, user.LastName).Scan(&userIdStr)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("postgres: create user: %w", err)
	}
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("postgres: create user while parsing id: %w", err)
	}
	return userId, nil
}

func (r *PostgresSQLDB) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `SELECT id, email, password,first_name,last_name FROM users WHERE email = $1`
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		return nil, fmt.Errorf("postgres:find user by email: %w", err)
	}
	return &user, nil
}

func (r *PostgresSQLDB) FindUserById(ctx context.Context, id uuid.UUID) (*models.User, error) {

	var user models.User
	query := `SELECT * FROM users WHERE id = $1`
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		return nil, fmt.Errorf("postgres:find user by id %s: %w", id, err)
	}
	return &user, nil
}

func (r *PostgresSQLDB) FindAllUsers(ctx context.Context) ([]models.User, error) {
	var users []models.User

	query := `SELECT * FROM users`

	err := r.db.SelectContext(ctx, &users, query)
	if err != nil {
		return nil, fmt.Errorf("postgres:find all users : %w", err)
	}

	return users, nil
}
