package controller

import (
	"errors"
	"time"

	"github.com/1akhilpandey/go-messaging/db"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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

// LoginUser authenticates a user and returns a JWT token.
func LoginUser(email, password string) (string, error) {
	user, err := db.GetUserByEmail(email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// Verify password using bcrypt (assumes user.Password is stored as a bcrypt hash)
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	// Create JWT token with a 72-hour expiration
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(72 * time.Hour).Unix(),
	})

	// Sign the token with a secret key (should be in configuration)
	tokenString, err := token.SignedString([]byte("mysecret"))
	if err != nil {
		return "", err
	}
	tokenID := uuid.New().String()
	expiresAt := time.Now().Add(72 * time.Hour)
	if err := db.InsertUserToken(tokenID, user.ID, tokenString, expiresAt); err != nil {
		return "", err
	}
	return tokenString, nil
}

// LogoutUser handles user logout.
// Revokes the provided token during security events such as logout.
func LogoutUser(token string) error {
	return db.RevokeToken(token)
}
