package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"mygochat/internal/chat"
)

func Connect() (*sql.DB, error) {
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// SaveMessage saves a message to the database
func SaveMessage(db *sql.DB, msg chat.Message) error {
	_, err := db.Exec(
		"INSERT INTO messages (username, message, time, color) VALUES (?, ?, ?, ?)",
		msg.Username, msg.Message, msg.Time, msg.Color,
	)
	if err != nil {
		log.Println("Failed to execute database query:", err)
		return err
	}
	return nil
}

// LoadMessages loads the last n messages from the database
func LoadMessages(db *sql.DB, limit int) ([]chat.Message, error) {
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

	messages := make([]chat.Message, 0)
	for rows.Next() {
		var msg chat.Message
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

	return messages, nil
}
