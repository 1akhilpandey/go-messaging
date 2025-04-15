package db

import (
	"fmt"
	"strconv"
	"time"
)

// Message represents a chat message.
type Message struct {
	ID        int64     `db:"id" json:"id"`
	ChatID    int64     `db:"chat_id" json:"chat_id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	Content   string    `db:"content" json:"content"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// InsertMessage saves a new message to the database.
// It assumes message.ChatID and message.UserID are already set correctly.
func InsertMessage(message *Message) error {
	query := `INSERT INTO messages (chat_id, user_id, content, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?)`
	// Using Go time for timestamps for clarity, though DB defaults could also be used.
	now := time.Now()
	result, err := DB.Exec(query, message.ChatID, message.UserID, message.Content, now, now)
	if err != nil {
		return fmt.Errorf("failed to insert message: %w", err)
	}

	// Retrieve and set the message.ID after insertion
	id, err := result.LastInsertId()
	if err != nil {
		// Log the error but don't necessarily fail the operation if ID retrieval fails
		fmt.Printf("Warning: failed to retrieve last insert ID for message: %v\n", err)
		// Depending on requirements, you might return nil here anyway,
		// or return a specific error indicating ID retrieval failure.
		// For now, we'll proceed without the ID if retrieval fails.
		return nil // Indicate overall success despite ID issue
	}
	message.ID = id // Set the ID back on the passed struct pointer

	return nil
}

// GetMessagesByChatID retrieves all messages for a specific chat ID.
// Messages are ordered by creation time (oldest first).
func GetMessagesByChatID(chatID string) ([]*Message, error) {
	// Convert string chat ID to int64
	chatIDInt, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid chat ID format: %w", err)
	}

	// Query to get all messages for the chat, ordered by creation time
	query := `SELECT id, chat_id, user_id, content, created_at, updated_at
			  FROM messages
			  WHERE chat_id = ?
			  ORDER BY created_at ASC`

	rows, err := DB.Query(query, chatIDInt)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.ChatID, &msg.UserID, &msg.Content, &msg.CreatedAt, &msg.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message row: %w", err)
		}
		messages = append(messages, &msg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating message rows: %w", err)
	}

	return messages, nil
}
