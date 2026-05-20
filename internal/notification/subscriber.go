package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	notificationdto "github.com/rizky/smart-grant/internal/notification/dto"
	"github.com/rs/zerolog/log"
)

const (
	stream          = "notifications"
	group           = "notification-workers"
	maxDelivery     = 3
	pendingInterval = 30 * time.Second
)

type Subscriber struct {
	rdb      *redis.Client
	pool     *pgxpool.Pool
	consumer string
}

func NewSubscriber(rdb *redis.Client, pool *pgxpool.Pool) *Subscriber {
	host, _ := os.Hostname()
	consumer := fmt.Sprintf("%s-%d", host, time.Now().UnixNano()%100000)
	return &Subscriber{rdb: rdb, pool: pool, consumer: consumer}
}

func (s *Subscriber) Run(ctx context.Context) error {
	if s.rdb == nil {
		log.Warn().Msg("Redis not available, subscriber skipped")
		return nil
	}

	if err := s.rdb.XGroupCreateMkStream(ctx, stream, group, "$").Err(); err != nil {
		if err.Error() != "BUSYGROUP Consumer Group name already exists" {
			return fmt.Errorf("create consumer group: %w", err)
		}
	}

	pendingTicker := time.NewTicker(pendingInterval)
	defer pendingTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-pendingTicker.C:
			s.processPending(ctx)
		default:
			s.processMessages(ctx)
		}
	}
}

func (s *Subscriber) processMessages(ctx context.Context) {
	results, err := s.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: s.consumer,
		Streams:  []string{stream, ">"},
		Count:    10,
		Block:    2 * time.Second,
	}).Result()
	if err != nil {
		if err == redis.Nil {
			return
		}
		log.Warn().Err(err).Msg("XReadGroup error, backing off")
		time.Sleep(time.Second)
		return
	}

	for _, result := range results {
		for _, message := range result.Messages {
			s.handleMessage(ctx, message)
		}
	}
}

func (s *Subscriber) processPending(ctx context.Context) {
	pending, err := s.rdb.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream:   stream,
		Group:    group,
		Start:    "-",
		End:      "+",
		Count:    10,
	}).Result()
	if err != nil {
		log.Warn().Err(err).Msg("XPendingExt error")
		return
	}

	for _, p := range pending {
		if p.RetryCount < maxDelivery {
			msg, err := s.rdb.XClaim(ctx, &redis.XClaimArgs{
				Stream:   stream,
				Group:    group,
				Consumer: s.consumer,
				MinIdle:  time.Duration(p.RetryCount+1) * 30 * time.Second,
				Messages: []string{p.ID},
			}).Result()
			if err != nil {
				continue
			}
			for _, m := range msg {
				s.handleMessage(ctx, m)
			}
		} else {
			log.Error().
				Str("msg_id", p.ID).
				Int("retries", int(p.RetryCount)).
				Msg("message exceeded max delivery attempts")
			s.rdb.XAck(ctx, stream, group, p.ID)
		}
	}
}

func (s *Subscriber) handleMessage(ctx context.Context, msg redis.XMessage) {
	payload, ok := msg.Values["payload"].(string)
	if !ok {
		s.acknowledge(ctx, msg.ID)
		return
	}

	var event notificationdto.NotificationEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		log.Warn().Err(err).Str("msg_id", msg.ID).Msg("failed to unmarshal notification payload")
		s.acknowledge(ctx, msg.ID)
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

	s.acknowledge(ctx, msg.ID)
}

func (s *Subscriber) acknowledge(ctx context.Context, msgID string) {
	if err := s.rdb.XAck(ctx, stream, group, msgID).Err(); err != nil {
		log.Warn().Err(err).Str("msg_id", msgID).Msg("XAck failed")
	}
}
