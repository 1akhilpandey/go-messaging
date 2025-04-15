package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/1akhilpandey/go-messaging/app/middleware"
	"github.com/1akhilpandey/go-messaging/db"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client represents a middleman between the websocket connection and the hub.
type Client struct {
	Hub  *Hub
	Conn *websocket.Conn
	// Buffered channel of outbound messages.
	Send chan []byte
	// UserID associated with this client (Needs to be populated on authentication/connection)
	UserID int64
	// ChatID this client is associated with (Needs to be populated on joining a chat/connection)
	ChatID int64
}

// ServeWs handles websocket requests from the peer.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// Extract username from the request context (set by AuthMiddleware)
	username, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok {
		log.Println("ServeWs: Missing username in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user from database
	user, err := db.GetUserByUsername(username)
	if err != nil {
		log.Printf("ServeWs: Error getting user by username %s: %v", username, err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Get chat ID from query parameter
	chatIDStr := r.URL.Query().Get("chat_id")
	if chatIDStr == "" {
		log.Println("ServeWs: Missing chat_id query parameter")
		http.Error(w, "Missing chat_id parameter", http.StatusBadRequest)
		return
	}

	// Convert user ID and chat ID to int64
	userID, err := strconv.ParseInt(user.ID, 10, 64)
	if err != nil {
		log.Printf("ServeWs: Error parsing user ID %s: %v", user.ID, err)
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		log.Printf("ServeWs: Error parsing chat ID %s: %v", chatIDStr, err)
		http.Error(w, "Invalid chat ID", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("ServeWs upgrade error:", err)
		return
	}

	client := &Client{
		Hub:    hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		UserID: userID,
		ChatID: chatID,
	}
	client.Hub.Register <- client

	// Start reading and writing pumps for the client.
	go client.writePump()
	go client.readPump()
}

// WsMessage defines the expected structure for incoming websocket messages.
type WsMessage struct {
	Content string `json:"content"`
	ChatID  int64  `json:"chat_id,omitempty"`
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("readPump error: %v", err)
			}
			break
		}
		// Attempt to parse the message and persist it
		var wsMsg WsMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("readPump: Error unmarshalling message: %v", err)
		} else {
			// Persist the message if UserID and ChatID are set
			if c.UserID != 0 && c.ChatID != 0 {
				dbMsg := &db.Message{
					ChatID:  c.ChatID,
					UserID:  c.UserID,
					Content: wsMsg.Content,
					// CreatedAt/UpdatedAt handled by db.InsertMessage or DB defaults
				}
				if err := db.InsertMessage(dbMsg); err != nil {
					log.Printf("readPump: Error inserting message: %v", err)
				}
			} else {
				// Log if context is missing
				log.Printf("Skipping persistence due to missing context (User: %d, Chat: %d)", c.UserID, c.ChatID)
			}
		}

		// Add the chat ID to the message before broadcasting
		var broadcastMsg WsMessage
		if err := json.Unmarshal(message, &broadcastMsg); err != nil {
			log.Printf("readPump: Error unmarshalling message for broadcast: %v", err)
			// Still broadcast the original message if we can't unmarshal it
			c.Hub.Broadcast <- message
		} else {
			// Set the chat ID in the message
			broadcastMsg.ChatID = c.ChatID

			// Marshal the message back to bytes
			modifiedMessage, err := json.Marshal(broadcastMsg)
			if err != nil {
				log.Printf("readPump: Error marshalling message with chat ID: %v", err)
				// Fallback to original message
				c.Hub.Broadcast <- message
			} else {
				// Broadcast the modified message with chat ID
				c.Hub.Broadcast <- modifiedMessage
			}
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Check if this message is intended for this client's chat
			var wsMsg WsMessage
			if err := json.Unmarshal(message, &wsMsg); err != nil {
				// If we can't unmarshal the message, we can't determine the chat ID
				// For backward compatibility, we'll still send the message
				log.Printf("writePump: Error unmarshalling message: %v", err)
			} else if wsMsg.ChatID != 0 && wsMsg.ChatID != c.ChatID {
				// Skip messages not intended for this chat
				continue
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message, but only if they match the chat ID
			n := len(c.Send)
			for i := 0; i < n; i++ {
				queuedMsg := <-c.Send

				// Check if this queued message is for this chat
				var queuedWsMsg WsMessage
				shouldSend := true
				if err := json.Unmarshal(queuedMsg, &queuedWsMsg); err == nil {
					if queuedWsMsg.ChatID != 0 && queuedWsMsg.ChatID != c.ChatID {
						shouldSend = false
					}
				}

				if shouldSend {
					w.Write([]byte("\n"))
					w.Write(queuedMsg)
				}
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
