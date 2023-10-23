package main

import "testing"

func TestLoadMessages(t *testing.T) {
	// вызываем функцию loadMessages с лимитом 2
	messages, err := loadMessages(2)
	if err != nil {
		t.Fatalf("failed to load messages: %v", err)
	}

	// проверяем, что возвращено 2 сообщения
	if len(messages) != 2 {
		t.Errorf("expected %d messages, but got %d", 2, len(messages))
	}
}
