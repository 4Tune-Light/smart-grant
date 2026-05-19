package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type Subscriber struct {
	rdb  *redis.Client
	pool *pgxpool.Pool
}

func NewSubscriber(rdb *redis.Client, pool *pgxpool.Pool) *Subscriber {
	return &Subscriber{rdb: rdb, pool: pool}
}

func (s *Subscriber) Run(ctx context.Context) error {
	group := "notification-workers"
	stream := "notifications"

	if err := s.rdb.XGroupCreateMkStream(ctx, stream, group, "$").Err(); err != nil {
		if err.Error() != "BUSYGROUP Consumer Group name already exists" {
			return fmt.Errorf("create consumer group: %w", err)
		}
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			s.processMessages(ctx, stream, group)
		}
	}
}

func (s *Subscriber) processMessages(ctx context.Context, stream string, group string) {
	results, err := s.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: "worker-1",
		Streams:  []string{stream, ">"},
		Count:    10,
		Block:    2 * time.Second,
	}).Result()
	if err != nil {
		return
	}

	for _, result := range results {
		for _, message := range result.Messages {
			s.handleMessage(ctx, stream, group, message)
		}
	}
}

func (s *Subscriber) handleMessage(ctx context.Context, stream string, group string, msg redis.XMessage) {
	payload, ok := msg.Values["payload"].(string)
	if !ok {
		s.acknowledge(ctx, stream, group, msg.ID)
		return
	}

	var event NotificationEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		log.Warn().Err(err).Str("msg_id", msg.ID).Msg("failed to unmarshal notification payload")
		s.acknowledge(ctx, stream, group, msg.ID)
		return
	}

	n := &Notification{
		UserID: event.UserID,
		Type:   event.Type,
		Title:  event.Title,
		Body:   event.Body,
	}

	if err := s.pool.QueryRow(ctx,
		`INSERT INTO notifications (user_id, type, title, body) VALUES ($1, $2, $3, $4) RETURNING id, created_at`,
		n.UserID, n.Type, n.Title, n.Body,
	).Scan(&n.ID, &n.CreatedAt); err != nil {
		log.Error().Err(err).Str("msg_id", msg.ID).Msg("failed to save notification to DB")
		return
	}

	event.ID = n.ID
	event.CreatedAt = n.CreatedAt.Format(time.RFC3339)
	payloadBytes, _ := json.Marshal(event)

	if err := s.rdb.Publish(ctx, fmt.Sprintf("user:%s:notifications", event.UserID), string(payloadBytes)).Err(); err != nil {
		log.Warn().Err(err).Str("user_id", event.UserID).Msg("failed to publish notification to pubsub")
	}

	s.acknowledge(ctx, stream, group, msg.ID)
}

func (s *Subscriber) acknowledge(ctx context.Context, stream string, group string, msgID string) {
	s.rdb.XAck(ctx, stream, group, msgID)
}
