package chat

import (
	"context"
	"time"

	"mygochat/internal/censor"
	"mygochat/internal/model"
	"mygochat/internal/pkg/utils"
	"mygochat/internal/repository/message"
)

// Service holds chat business logic: persistence, censoring and broadcast.
type Service struct {
	repo         message.Repository
	censor       *censor.Censor
	hub          *Hub
	historyLimit int
}

func NewService(repo message.Repository, c *censor.Censor, hub *Hub, historyLimit int) *Service {
	return &Service{
		repo:         repo,
		censor:       c,
		hub:          hub,
		historyLimit: historyLimit,
	}
}

// History returns recent messages, censored, for a freshly connected client.
func (s *Service) History(ctx context.Context) ([]model.Message, error) {
	messages, err := s.repo.LoadRecent(ctx, s.historyLimit)
	if err != nil {
		return nil, err
	}
	for i := range messages {
		messages[i].Username = s.censor.Clean(messages[i].Username)
		messages[i].Message = s.censor.Clean(messages[i].Message)
	}
	return messages, nil
}

// Publish persists an incoming message and broadcasts a censored copy to all clients.
func (s *Service) Publish(ctx context.Context, msg model.Message) error {
	if msg.Color == "" {
		msg.Color = utils.GetRandomColor()
	}

	if err := s.repo.Save(ctx, &msg); err != nil {
		return err
	}

	msg.Username = s.censor.Clean(msg.Username)
	msg.Message = s.censor.Clean(msg.Message)
	s.hub.Broadcast(msg)
	return nil
}

// SystemMessage builds a server-authored message (e.g. the welcome banner).
func SystemMessage(text string) model.Message {
	return model.Message{
		Username: "Server",
		Message:  text,
		Time:     time.Now(),
		Color:    utils.GetRandomColor(),
	}
}
