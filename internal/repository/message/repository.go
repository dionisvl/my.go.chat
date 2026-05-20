package message

import (
	"context"

	"mygochat/internal/model"
	"mygochat/internal/platform/dbtx"
)

type Repository interface {
	// Save persists a message and fills its generated ID.
	Save(ctx context.Context, msg *model.Message) error
	// LoadRecent returns up to limit most recent messages in chronological (oldest-first) order.
	LoadRecent(ctx context.Context, limit int) ([]model.Message, error)
}

type repository struct {
	db dbtx.DB
}

func New(db dbtx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Save(ctx context.Context, msg *model.Message) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO messages (username, message, color)
		 VALUES ($1, $2, $3)
		 RETURNING id, created_at`,
		msg.Username, msg.Message, msg.Color,
	).Scan(&msg.ID, &msg.Time)
}

func (r *repository) LoadRecent(ctx context.Context, limit int) ([]model.Message, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, username, message, created_at, color
		 FROM (
		     SELECT id, username, message, created_at, color
		     FROM messages
		     ORDER BY created_at DESC, id DESC
		     LIMIT $1
		 ) recent
		 ORDER BY created_at ASC, id ASC`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]model.Message, 0, limit)
	for rows.Next() {
		var m model.Message
		if err := rows.Scan(&m.ID, &m.Username, &m.Message, &m.Time, &m.Color); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}
