package controller

import (
	"errors"

	"example.com/go-messaging/db"
)

// CreateUserInput represents the data required to create a user.
type CreateUserInput struct {
	Username string
	Email    string
	Password string
}

// UserResponse represents the data returned after creating or retrieving a user.
type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// CreateUser processes the creation of a new user and interacts with the database.
func CreateUser(input CreateUserInput) (UserResponse, error) {
	if input.Username == "" {
		return UserResponse{}, errors.New("username is required")
	}

	newUser, err := db.InsertUser(input.Username, input.Email, input.Password)
	if err != nil {
		return UserResponse{}, err
	}

	return UserResponse{
		ID:       newUser.ID,
		Username: newUser.Username,
		Email:    newUser.Email,
	}, nil
}

// GetUser retrieves a user by ID from the database.
func GetUser(id string) (UserResponse, error) {
	user, err := db.GetUserByID(id)
	if err != nil {
		return UserResponse{}, err
	}

	return UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}, nil
}
