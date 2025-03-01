package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"mygochat/internal/chat"
	"mygochat/internal/database"
	"mygochat/internal/handler"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	log.Println("Starting server... ok")

	// Initialize database
	db, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)

	// Load profanities
	chat.LoadProfanities(os.Getenv("PROFANITIES"))

	// Setup routes
	http.HandleFunc("/", handler.HandleRoot)
	http.HandleFunc("/ws", handler.HandleWebSocket(db))

	log.Println("Starting ListenAndServe route...")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
