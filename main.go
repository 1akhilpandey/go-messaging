package main

import (
	"log"
	"net/http"

	"github.com/1akhilpandey/go-messaging/app/api/handler"
	"github.com/1akhilpandey/go-messaging/app/ws"
	"github.com/1akhilpandey/go-messaging/db"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Initialize the database
	database, err := db.InitDB("chatapp.db")
	if err != nil {
		log.Fatalf("Could not initialize database: %v", err)
	}
	defer database.Close()

	// Create a new WebSocket hub and run it
	hub := ws.NewHub()
	go hub.Run()

	// Set up router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// TODO: Add session middleware here if implemented

	// User routes
	r.Route("/user", func(r chi.Router) {
		r.Post("/signup", handler.CreateUserHandler)
		r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "login not implemented", http.StatusNotImplemented)
		})
		r.Post("/logout", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "logout not implemented", http.StatusNotImplemented)
		})
	})

	// Chat routes
	r.Route("/chat", func(r chi.Router) {
		r.Get("/messages", handler.GetChatHandler)
		r.Post("/message", handler.CreateChatHandler)
		// Group chat endpoints not implemented
		//r.Post("/group", handler.CreateGroup)
		//r.Get("/group/{groupID}", handler.GetGroup)
	})

	// WebSocket endpoint
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hub, w, r)
	})

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}
