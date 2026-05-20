package model

import "time"

// Message is a single chat message as stored and broadcast.
type Message struct {
	ID       int64     `json:"id"`
	Username string    `json:"username"`
	Message  string    `json:"message"`
	Time     time.Time `json:"time"`
	Color    string    `json:"color"`
}
