package db

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/1akhilpandey/go-messaging/migrate"
	_ "github.com/mattn/go-sqlite3"
)

// GetUserByEmail retrieves a user by email from the database.
func GetUserByEmail(email string) (*User, error) {
	row := DB.QueryRow("SELECT id, username, email, password FROM users WHERE email = ?", email)
	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

var DB *sql.DB

// InitDB initializes a new SQLite database connection and assigns it to the package-level DB variable.
func InitDB(filepath string) (*sql.DB, error) {
	var err error
	DB, err = sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}
	if err = DB.Ping(); err != nil {
		return nil, err
	}
	return DB, nil
}

type User struct {
	ID       string
	Username string
	Email    string
	Password string
}

func InsertUser(username, email, password string) (*User, error) {
	// Validate input parameters
	if username == "" || email == "" || password == "" {
		return nil, fmt.Errorf("username, email, and password must not be empty")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Prepare the SQL query.
	query := "INSERT INTO users (username, email, password) VALUES (?, ?, ?)"
	result, err := DB.Exec(query, username, email, string(hashedPassword))
	if err != nil {
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}

	// Retrieve the ID of the newly inserted user.
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve last inserted ID: %w", err)
	}

	return &User{
		ID:       strconv.FormatInt(id, 10),
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
	}, nil
}

func GetUserByID(id string) (*User, error) {
	row := DB.QueryRow("SELECT id, username, email, password FROM users WHERE id = ?", id)
	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

type Chat struct {
	ID      string
	Title   string
	UserIDs []string
	IsGroup bool
}

func InsertChat(title string, userIDs []string, isGroup bool) (*Chat, error) {
	userIDsStr := strings.Join(userIDs, ",")
	res, err := DB.Exec("INSERT INTO chats (title, user_ids, is_group) VALUES (?, ?, ?)", title, userIDsStr, isGroup)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &Chat{
		ID:      strconv.FormatInt(id, 10),
		Title:   title,
		UserIDs: userIDs,
		IsGroup: isGroup,
	}, nil
}

func GetChatByID(id string) (*Chat, error) {
	row := DB.QueryRow("SELECT id, title, user_ids, is_group FROM chats WHERE id = ?", id)
	var chat Chat
	var userIDsStr string
	err := row.Scan(&chat.ID, &chat.Title, &userIDsStr, &chat.IsGroup)
	if err != nil {
		return nil, err
	}
	if userIDsStr != "" {
		chat.UserIDs = strings.Split(userIDsStr, ",")
	} else {
		chat.UserIDs = []string{}
	}
	return &chat, nil
}

// SetupDatabase initializes the database connection and applies migrations.
// It returns the database connection or an error if any step fails.
func SetupDatabase(filepath string) (*sql.DB, error) {
	dbConn, err := InitDB(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	migrationURL := fmt.Sprintf("sqlite3://%s", filepath)
	if err := migrate.Migrate(migrationURL); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}
	return dbConn, nil
}

func InsertUserToken(tokenID, userID, tokenValue string, expiresAt time.Time) error {
	query := "INSERT INTO user_tokens (token_id, user_id, token_value, expires_at) VALUES (?, ?, ?, ?)"
	_, err := DB.Exec(query, tokenID, userID, tokenValue, expiresAt)
	return err
}

func RevokeToken(token string) error {
	query := "DELETE FROM user_tokens WHERE token_value = ?"
	_, err := DB.Exec(query, token)
	return err
}

// GetUserByUsername retrieves a user by username from the database.
func GetUserByUsername(username string) (*User, error) {
	row := DB.QueryRow("SELECT id, username, email, password FROM users WHERE username = ?", username)
	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetChatsByUserID retrieves all chats where the specified user is a participant.
func GetChatsByUserID(userID string) ([]*Chat, error) {
	// In our database, user_ids is a comma-separated list of user IDs
	// We need to find chats where the user's ID is in this list
	rows, err := DB.Query("SELECT id, title, user_ids, is_group FROM chats WHERE user_ids LIKE ?", "%"+userID+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []*Chat
	for rows.Next() {
		var chat Chat
		var userIDsStr string
		err := rows.Scan(&chat.ID, &chat.Title, &userIDsStr, &chat.IsGroup)
		if err != nil {
			return nil, err
		}

		// Parse the comma-separated user IDs
		if userIDsStr != "" {
			chat.UserIDs = strings.Split(userIDsStr, ",")
		} else {
			chat.UserIDs = []string{}
		}

		// Verify that the user is actually in the participants list
		// This is needed because our LIKE query might match substrings
		for _, id := range chat.UserIDs {
			if id == userID {
				chats = append(chats, &chat)
				break
			}
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return chats, nil
}
