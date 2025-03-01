package main

import (
	"database/sql"
	"fmt"
	"github.com/TwiN/go-away"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Message struct {
	Id       int64     `json:"id"`
	Username string    `json:"username"`
	Message  string    `json:"message"`
	Time     time.Time `json:"time" db:"time" sql:"type:datetime"`
	Color    string    `json:"color" db:"color" sql:"type:string"`
}

var db *sql.DB
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func getRandomColor() string {
	rand.Seed(time.Now().UnixNano())
	letters := "6789ABCDEF"
	color := "#"
	for i := 0; i < 6; i++ {
		color += string(letters[rand.Intn(len(letters))])
	}
	return color
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
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	log.Println("Starting server... ok")
	log.Println("Starting root route...")
	http.HandleFunc("/", handleRoot)
	log.Println("... ok")
	log.Println("Starting ws route...")
	http.HandleFunc("/ws", handleWebSocket)
	log.Println("... ok")

	log.Println("Starting mysql logic...")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Println("... ok")

	// Load profanities from .env
	loadProfanities()

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

	// Load the last 50 messages from the database
	messages, err := loadMessages(50)
	if err != nil {
		log.Println("Failed to load messages:", err)
		return
	}
	sendMessages(messages, conn)
	sendWelcome(conn)

	// Listen messages from the WebSocket connection
	for {
		// Read a message from the WebSocket connection
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Println("Client disconnected normally")
			} else {
				log.Println("Failed to read message:", err)
			}
			delete(clients, conn)
			break
		}

		// Log the message received from the client
		log.Printf("Received message from %s: %s", msg.Username, msg.Message)

		// Add a timestamp to the message
		msg.Time = time.Now().Local()

		// Save the message to the database
		err = saveMessage(msg)
		if err != nil {
			log.Println("Failed to save message:", err)
			break
		}

		// Broadcast the message to all connected clients
		for client := range clients {
			msg.Username = censor(msg.Username)
			msg.Message = censor(msg.Message)

			err = client.WriteJSON(msg)
			if err != nil {
				log.Println("Failed to broadcast message:", err)
				delete(clients, client)
			}
		}
	}
}

// Initialize profanity detector
var profanityDetector *goaway.ProfanityDetector

func loadProfanities() {
	// Get profanities from environment variable
	profanities := os.Getenv("PROFANITIES")
	log.Println("PROFANITIES: ")
	log.Println(profanities)
	customProfanities := strings.Split(profanities, ",")
	// Create a custom profanity detector
	profanityDetector = goaway.NewProfanityDetector().WithCustomDictionary(customProfanities, nil, nil)
}

func censor(str string) string {
	// Replace specific words and use the profanity detector
	str = strings.ReplaceAll(str, "хуй", "***")
	return profanityDetector.Censor(str)
}

// Send the message list to the client
func sendMessages(messages []Message, conn *websocket.Conn) {
	for _, msg := range messages {
		msg.Username = censor(msg.Username)
		msg.Message = censor(msg.Message)

		err := conn.WriteJSON(msg)
		if err != nil {
			log.Println("Failed to send message:", err)
			delete(clients, conn)
			break
		}
	}
}

// Send special welcome text with timeout
func sendWelcome(conn *websocket.Conn) {
	welcomeMessageText := os.Getenv("WELCOME_MESSAGE")
	welcomeMsg := Message{
		Username: "Golang Server",
		Message:  welcomeMessageText,
		Time:     time.Now().Local(),
		Color:    getRandomColor(),
	}

	// Get the welcome timeout from the environment
	welcomeTimeout, err := strconv.Atoi(os.Getenv("WELCOME_TIMEOUT"))
	if err != nil {
		log.Println("Failed to parse WELCOME_TIMEOUT:", err)
		return
	}

	// Schedule the welcome message to be sent after the welcome timeout
	time.AfterFunc(time.Duration(welcomeTimeout)*time.Second, func() {
		messages := make([]Message, 0)
		messages = append(messages, welcomeMsg)
		sendMessages(messages, conn)
	})
}

func saveMessage(msg Message) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}
	// Insert the message into the database
	_, err := db.Exec("INSERT INTO messages (username, message, time, color) VALUES (?, ?, ?, ?)", msg.Username, msg.Message, msg.Time, msg.Color)
	if err != nil {
		log.Println("Failed to execute database query:", err)
		return err
	}
	return nil
}

// Load the last `limit` messages from the database
func loadMessages(limit int) ([]Message, error) {
	// Query the database for the last `limit` messages
	query := fmt.Sprintf(`
	SELECT * FROM (
		SELECT id, username, message, time, color
		FROM messages
		ORDER BY time DESC
		LIMIT %d
	) AS sub
	ORDER BY time ASC
`, limit)

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate over the rows and create a slice of messages
	messages := make([]Message, 0)
	for rows.Next() {
		var msg Message
		var timeStr string
		err := rows.Scan(&msg.Id, &msg.Username, &msg.Message, &timeStr, &msg.Color)
		if err != nil {
			return nil, err
		}

		msg.Time, err = time.ParseInLocation("2006-01-02 15:04:05", timeStr, time.Local)
		if err != nil {
			return nil, err
		}

		messages = append(messages, msg)
	}

	// Return the slice of messages
	return messages, nil
}
