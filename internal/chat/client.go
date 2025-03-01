package chat

import (
	"log"
	"mygochat/internal/pkg/utils"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// SendMessages sends messages to a specific client
func SendMessages(messages []Message, conn *websocket.Conn) {
	for _, msg := range messages {
		msg.Username = Censor(msg.Username)
		msg.Message = Censor(msg.Message)

		err := conn.WriteJSON(msg)
		if err != nil {
			log.Println("Failed to send message:", err)
			delete(Clients, conn)
			break
		}
	}
}

// SendWelcome sends welcome message with a delay
func SendWelcome(conn *websocket.Conn, welcomeMessage string, timeout int) {
	welcomeMsg := Message{
		Username: "Server",
		Message:  welcomeMessage,
		Time:     time.Now().Local(),
		Color:    utils.GetRandomColor(),
	}

	// Schedule the welcome message
	time.AfterFunc(time.Duration(timeout)*time.Second, func() {
		messages := []Message{welcomeMsg}
		SendMessages(messages, conn)
	})
}

// BroadcastMessage sends a message to all connected clients
func BroadcastMessage(msg Message) {
	for client := range Clients {
		msg.Username = Censor(msg.Username)
		msg.Message = Censor(msg.Message)

		err := client.WriteJSON(msg)
		if err != nil {
			log.Println("Failed to broadcast message:", err)
			delete(Clients, client)
		}
	}
}
