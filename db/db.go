package db

import (
	"database/sql"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

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
	res, err := DB.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)", username, email, password)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &User{
		ID:       strconv.FormatInt(id, 10),
		Username: username,
		Email:    email,
		Password: password,
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
