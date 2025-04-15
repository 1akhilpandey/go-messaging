package main

import (
	"log"
	"net/http"

	"github.com/1akhilpandey/go-messaging/app/api/handler"
	authMiddleware "github.com/1akhilpandey/go-messaging/app/middleware"
	"github.com/1akhilpandey/go-messaging/app/ws"
	"github.com/1akhilpandey/go-messaging/db"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Set up the database and apply migrations.
	database, err := db.SetupDatabase("chatapp.db")
	if err != nil {
		log.Fatalf("Database setup failed: %v", err)
	}
	defer database.Close()

	// Create a new WebSocket hub and run it.
	hub := ws.NewHub()
	go hub.Run()

	// Set up router.
	r := chi.NewRouter()
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	// User routes.
	r.Route("/user", func(r chi.Router) {
		// Public endpoints.
		r.Post("/signup", handler.CreateUserHandler)
		r.Post("/login", handler.LoginUserHandler)

		// Protected endpoint.
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.AuthMiddleware)
			r.Post("/logout", handler.LogoutUserHandler)
		})
	})

	// Protected routes.
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.AuthMiddleware)

		// Chat routes.
		r.Route("/chat", func(r chi.Router) {
			r.Get("/messages/{id}", handler.GetChatHandler)
			r.Post("/message", handler.CreateChatHandler)
			r.Get("/user", handler.GetUserChatsHandler)
		})

		// WebSocket endpoint.
		r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			ws.ServeWs(hub, w, r)
		})
	})

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}
