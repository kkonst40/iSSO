package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kkonst40/sso-service/internal/config"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Producer struct {
	client *kgo.Client
}

const topicUserEvents = "user-events"

func NewProducer(cfg *config.Config) (*Producer, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(fmt.Sprintf("%s:%s", cfg.Kafka.Host, cfg.Kafka.Port)),
		kgo.DefaultProduceTopic(topicUserEvents),
		kgo.AllowAutoTopicCreation(),
	)
	if err != nil {
		return nil, err
	}

	return &Producer{
		client: cl,
	}, nil
}

func (p *Producer) SendLoginUpdate(ctx context.Context, userID uuid.UUID, login string) error {
	payloadBytes, _ := json.Marshal(loginUpdatePayload{UserID: userID, Login: login})

	event := eventMessage{
		Type:      eventTypeLoginUpdate,
		Payload:   payloadBytes,
		CreatedAt: time.Now(),
	}

	key := fmt.Appendf(nil, "user_%v", userID)
	val, _ := json.Marshal(event)

	return p.client.ProduceSync(ctx, &kgo.Record{
		Key:   key,
		Value: val,
	}).FirstErr()
}

func (p *Producer) SendSessionInvalidation(ctx context.Context, sessionID uuid.UUID, ttlDays int) error {
	payloadBytes, _ := json.Marshal(sessionInvalidationPayload{SessionID: sessionID, TTLDays: ttlDays})

	event := eventMessage{
		Type:      eventTypeSessionInvalidation,
		Payload:   payloadBytes,
		CreatedAt: time.Now(),
	}

	val, _ := json.Marshal(event)
	key := fmt.Appendf(nil, "session_%v", sessionID)

	return p.client.ProduceSync(ctx, &kgo.Record{
		Key:   key,
		Value: val,
	}).FirstErr()
}

func (p *Producer) Close() {
	p.client.Close()
}
