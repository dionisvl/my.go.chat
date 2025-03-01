package handler

import (
	"database/sql"
	"log"
	"mygochat/internal/pkg/utils"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"mygochat/internal/chat"
	"mygochat/internal/database"
)

func HandleWebSocket(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Upgrade the HTTP connection to a WebSocket connection
		conn, err := chat.Upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Failed to upgrade connection:", err)
			return
		}
		defer conn.Close()

		// Add the client to the list of connected clients
		chat.Clients[conn] = true

		// Load the last 50 messages from the database
		messages, err := database.LoadMessages(db, 50)
		if err != nil {
			log.Println("Failed to load messages:", err)
			return
		}

		chat.SendMessages(messages, conn)

		// Get welcome message and timeout
		welcomeMessage := os.Getenv("WELCOME_MESSAGE")
		welcomeTimeout, _ := strconv.Atoi(os.Getenv("WELCOME_TIMEOUT"))
		chat.SendWelcome(conn, welcomeMessage, welcomeTimeout)

		// Listen for messages from the WebSocket connection
		for {
			var msg chat.Message
			err := conn.ReadJSON(&msg)
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					log.Println("Client disconnected normally")
				} else {
					log.Println("Failed to read message:", err)
				}
				delete(chat.Clients, conn)
				break
			}

			log.Printf("Received message from %s: %s", msg.Username, msg.Message)

			// Add timestamp and random color if not provided
			msg.Time = time.Now().Local()
			if msg.Color == "" {
				msg.Color = utils.GetRandomColor()
			}

			// Save the message to the database
			err = database.SaveMessage(db, msg)
			if err != nil {
				log.Println("Failed to save message:", err)
				break
			}

			// Broadcast the message to all connected clients
			chat.BroadcastMessage(msg)
		}
	}
}
