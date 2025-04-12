package controller

import (
	"errors"

	"example.com/go-messaging/db"
)

// CreateChatInput represents the data required to create a new chat.
type CreateChatInput struct {
	Title   string
	UserIDs []string
	IsGroup bool
}

// ChatResponse represents the data returned after creating or retrieving a chat.
type ChatResponse struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	UserIDs []string `json:"user_ids"`
	IsGroup bool     `json:"is_group"`
}

// CreateChat processes the creation of a new chat and interacts with the database.
func CreateChat(input CreateChatInput) (ChatResponse, error) {
	if input.Title == "" {
		return ChatResponse{}, errors.New("chat title is required")
	}
	chat, err := db.InsertChat(input.Title, input.UserIDs, input.IsGroup)
	if err != nil {
		return ChatResponse{}, err
	}
	return ChatResponse{
		ID:      chat.ID,
		Title:   chat.Title,
		UserIDs: chat.UserIDs,
		IsGroup: chat.IsGroup,
	}, nil
}

// GetChat retrieves a chat by ID from the database.
func GetChat(id string) (ChatResponse, error) {
	chat, err := db.GetChatByID(id)
	if err != nil {
		return ChatResponse{}, err
	}
	return ChatResponse{
		ID:      chat.ID,
		Title:   chat.Title,
		UserIDs: chat.UserIDs,
		IsGroup: chat.IsGroup,
	}, nil
}
