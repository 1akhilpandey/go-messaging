package controller

import (
	"errors"
	"strconv"
	"time"

	"github.com/1akhilpandey/go-messaging/db"
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

// MessageResponse represents a single message in the chat.
type MessageResponse struct {
	ID        string    `json:"id"`
	ChatID    string    `json:"chat_id"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetChatMessagesResponse represents the data returned when retrieving messages for a chat.
type GetChatMessagesResponse struct {
	Messages []MessageResponse `json:"messages"`
	Count    int               `json:"count"`
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

// GetChat retrieves all messages for a chat by ID from the database.
func GetChat(id string) (GetChatMessagesResponse, error) {
	// First verify that the chat exists
	_, err := db.GetChatByID(id)
	if err != nil {
		return GetChatMessagesResponse{}, err
	}

	// Get all messages for the chat
	messages, err := db.GetMessagesByChatID(id)
	if err != nil {
		return GetChatMessagesResponse{}, err
	}

	// Convert to response format
	var messageResponses []MessageResponse
	for _, msg := range messages {
		messageResponses = append(messageResponses, MessageResponse{
			ID:        strconv.FormatInt(msg.ID, 10),
			ChatID:    strconv.FormatInt(msg.ChatID, 10),
			UserID:    strconv.FormatInt(msg.UserID, 10),
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt,
			UpdatedAt: msg.UpdatedAt,
		})
	}

	return GetChatMessagesResponse{
		Messages: messageResponses,
		Count:    len(messageResponses),
	}, nil
}

// GetUserChatsResponse represents the data returned when retrieving a user's chats.
type GetUserChatsResponse struct {
	Chats []ChatResponse `json:"chats"`
	Count int            `json:"count"`
}

// GetUserChats retrieves all chats where the specified user is a participant.
func GetUserChats(username string) (GetUserChatsResponse, error) {
	// Get the user by username
	user, err := db.GetUserByUsername(username)
	if err != nil {
		return GetUserChatsResponse{}, err
	}

	// Get all chats for the user
	chats, err := db.GetChatsByUserID(user.ID)
	if err != nil {
		return GetUserChatsResponse{}, err
	}

	// Convert to response format
	var chatResponses []ChatResponse
	for _, chat := range chats {
		chatResponses = append(chatResponses, ChatResponse{
			ID:      chat.ID,
			Title:   chat.Title,
			UserIDs: chat.UserIDs,
			IsGroup: chat.IsGroup,
		})
	}

	return GetUserChatsResponse{
		Chats: chatResponses,
		Count: len(chatResponses),
	}, nil
}
