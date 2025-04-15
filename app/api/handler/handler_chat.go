package handler

import (
	"encoding/json"
	"net/http"

	"github.com/1akhilpandey/go-messaging/app/api/controller"
	"github.com/1akhilpandey/go-messaging/app/middleware"
	"github.com/go-chi/chi/v5"
)

// CreateChatRequest defines the expected payload for creating a new chat.
type CreateChatRequest struct {
	Title   string   `json:"title"`
	UserIDs []string `json:"user_ids"`
	IsGroup bool     `json:"is_group"`
}

// CreateChatHandler handles the HTTP POST request to create a new chat.
func CreateChatHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Build input for controller.
	input := controller.CreateChatInput{
		Title:   req.Title,
		UserIDs: req.UserIDs,
		IsGroup: req.IsGroup,
	}

	chat, err := controller.CreateChat(input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chat)
}

// GetChatHandler handles the HTTP GET request to retrieve a chat by ID.
func GetChatHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Chat ID is required", http.StatusBadRequest)
		return
	}

	chat, err := controller.GetChat(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chat)
}

// GetUserChatsHandler handles the HTTP GET request to retrieve all chats for the authenticated user.
func GetUserChatsHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the username from the request context
	username, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get all chats for the user
	response, err := controller.GetUserChats(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
