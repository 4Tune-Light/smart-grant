package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rizky/smart-grant/internal/middleware"
)

type Service interface {
	Send(ctx context.Context, userID string, notifType string, title string, body string) error
	List(ctx context.Context, limit int, page int) ([]NotificationResponse, int, error)
	MarkRead(ctx context.Context, notificationID string) error
	Subscribe(ctx context.Context) (<-chan NotificationEvent, error)
}

type service struct {
	repo   Repository
	rdb    *redis.Client // nil if Redis unavailable — falls back to DB-only
	pubsub *redis.PubSub
}

func NewService(repo Repository, rdb *redis.Client) Service {
	return &service{repo: repo, rdb: rdb}
}

func (s *service) Send(ctx context.Context, userID string, notifType string, title string, body string) error {
	n := &Notification{
		UserID: userID,
		Type:   notifType,
		Title:  title,
		Body:   body,
	}

	if err := s.repo.Insert(ctx, n); err != nil {
		return fmt.Errorf("save notification: %w", err)
	}

	if s.rdb != nil {
		event := NotificationEvent{
			ID:        n.ID,
			UserID:    userID,
			Type:      notifType,
			Title:     title,
			Body:      body,
			CreatedAt: n.CreatedAt.Format(time.RFC3339),
		}
		data, _ := json.Marshal(event)

		streamEntry := map[string]interface{}{
			"user_id": userID,
			"payload": string(data),
		}

		if err := s.rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: "notifications",
			Values: streamEntry,
		}).Err(); err != nil {
			return fmt.Errorf("publish to stream: %w", err)
		}
	}

	return nil
}

func (s *service) List(ctx context.Context, limit int, page int) ([]NotificationResponse, int, error) {
	userID, _ := ctx.Value(middleware.AuthUserIDKey).(string)
	offset := (page - 1) * limit

	notifications, total, err := s.repo.FindByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]NotificationResponse, len(notifications))
	for i, n := range notifications {
		responses[i] = NotificationResponse{
			ID:        n.ID,
			Type:      n.Type,
			Title:     n.Title,
			Body:      n.Body,
			IsRead:    n.IsRead,
			CreatedAt: n.CreatedAt,
		}
	}

	return responses, total, nil
}

func (s *service) MarkRead(ctx context.Context, notificationID string) error {
	userID, _ := ctx.Value(middleware.AuthUserIDKey).(string)
	return s.repo.MarkRead(ctx, notificationID, userID)
}

func (s *service) Subscribe(ctx context.Context) (<-chan NotificationEvent, error) {
	userID, _ := ctx.Value(middleware.AuthUserIDKey).(string)
	ch := make(chan NotificationEvent, 10)

	if s.rdb == nil {
		close(ch)
		return ch, nil
	}

	pubsub := s.rdb.Subscribe(ctx, fmt.Sprintf("user:%s:notifications", userID))

	go func() {
		defer pubsub.Close()
		defer close(ch)

		for {
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil {
				return
			}

			var event NotificationEvent
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				continue
			}

			select {
			case ch <- event:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}
