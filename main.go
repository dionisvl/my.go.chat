package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"time"
)

type Message struct {
	Id       int64  `json:"id"`
	Username string `json:"username"`
	Message  string `json:"message"`
	Time     int64  `json:"time"`
}

var db *sql.DB
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	// Read the contents of the index.html file
	data, err := os.ReadFile("web/index.html")
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	// Write the contents of the file to the response writer
	fmt.Fprintf(w, string(data))
}

func main() {
	err := os.Chdir("/go/src/app")
	log.Println("Starting server... ok")
	log.Println("Starting root route...")
	http.HandleFunc("/", handleRoot)
	log.Println("... ok")
	log.Println("Starting ws route...")
	http.HandleFunc("/ws", handleWebSocket)
	log.Println("... ok")

	log.Println("Starting mysql logic...")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	db, err = sql.Open("mysql", fmt.Sprintf("root:%s@tcp(%s:3306)/%s", dbPass, dbHost, dbName))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Println("... ok")

	log.Println("Starting ListenAndServe route...")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println("error during ListenAndServe:", err)
		log.Fatal(err)
		return
	}
	log.Println("... ok")

}

// Define a map to store all connected clients
var clients = make(map[*websocket.Conn]bool)

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}
	defer conn.Close()

	// Add the client to the list of connected clients
	clients[conn] = true

	// Send the welcome message to the client
	welcomeMessage := os.Getenv("WELCOME_MESSAGE")
	welcomeMsg := Message{
		Username: "GolangWebSocketServer",
		Message:  welcomeMessage,
		Time:     time.Now().Unix(),
	}
	err = conn.WriteJSON(welcomeMsg)
	if err != nil {
		log.Println("Failed to send message:", err)
		delete(clients, conn)
		return
	}

	// Load the last 50 messages from the database
	messages, err := loadMessages(15)
	if err != nil {
		log.Println("Failed to load messages:", err)
		return
	}

	// Send the messages to the client
	for _, msg := range messages {
		err = conn.WriteJSON(msg)
		if err != nil {
			log.Println("Failed to send message:", err)
			delete(clients, conn)
			break
		}
	}

	// Read messages from the WebSocket connection
	for {
		// Read a message from the WebSocket connection
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Failed to read message:", err)
			delete(clients, conn)
			break
		}

		// Log the message received from the client
		log.Printf("Received message from %s: %s", msg.Username, msg.Message)

		// Add a timestamp to the message
		msg.Time = time.Now().Unix()

		// Save the message to the database
		err = saveMessage(msg)
		if err != nil {
			log.Println("Failed to save message:", err)
			break
		}

		// Broadcast the message to all connected clients
		for client := range clients {
			err = client.WriteJSON(msg)
			if err != nil {
				log.Println("Failed to broadcast message:", err)
				delete(clients, client)
			}
		}
	}
}

func saveMessage(msg Message) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}
	// Insert the message into the database
	_, err := db.Exec("INSERT INTO messages (username, message, time) VALUES (?, ?, ?)", msg.Username, msg.Message, msg.Time)
	if err != nil {
		log.Println("Failed to execute database query:", err)
		return err
	}
	return nil
}

// Load the last `limit` messages from the database
func loadMessages(limit int) ([]Message, error) {
	// Query the database for the last `limit` messages
	rows, err := db.Query("SELECT * FROM messages ORDER BY time LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate over the rows and create a slice of messages
	messages := make([]Message, 0)
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.Id, &msg.Username, &msg.Message, &msg.Time)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	// Return the slice of messages
	return messages, nil
}
