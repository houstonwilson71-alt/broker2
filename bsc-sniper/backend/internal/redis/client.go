package redisclient

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	StreamNewPairs      = "stream:new_pairs"
	StreamApproved      = "stream:approved_tokens"
	PubSubEvents        = "pubsub:events"
	GroupFilter         = "group:filter"
	GroupExecutor       = "group:executor"
	ConsumerFilter      = "consumer:filter"
	ConsumerExecutor    = "consumer:executor"
)

type Client struct {
	rdb    *redis.Client
	logger *zap.Logger
}

func New(redisURL string, logger *zap.Logger) (*Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	opts.PoolSize = 20
	opts.MinIdleConns = 2
	opts.DialTimeout = 5 * time.Second
	opts.ReadTimeout = 3 * time.Second
	opts.WriteTimeout = 3 * time.Second

	rdb := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	c := &Client{rdb: rdb, logger: logger}
	if err := c.ensureConsumerGroups(context.Background()); err != nil {
		return nil, err
	}

	logger.Info("redis connected")
	return c, nil
}

func (c *Client) ensureConsumerGroups(ctx context.Context) error {
	streams := []struct {
		stream string
		group  string
	}{
		{StreamNewPairs, GroupFilter},
		{StreamApproved, GroupExecutor},
	}
	for _, s := range streams {
		err := c.rdb.XGroupCreateMkStream(ctx, s.stream, s.group, "0").Err()
		if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
			return fmt.Errorf("create group %s/%s: %w", s.stream, s.group, err)
		}
	}
	return nil
}

// PublishToStream publishes a JSON-encoded value to a Redis stream.
func (c *Client) PublishToStream(ctx context.Context, stream string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return c.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: map[string]interface{}{"data": string(data)},
	}).Err()
}

// ReadStream reads messages from a consumer group.
func (c *Client) ReadStream(ctx context.Context, stream, group, consumer string, count int64, block time.Duration) ([]redis.XStream, error) {
	return c.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  []string{stream, ">"},
		Count:    count,
		Block:    block,
	}).Result()
}

// AckMessage acknowledges a stream message.
func (c *Client) AckMessage(ctx context.Context, stream, group, id string) error {
	return c.rdb.XAck(ctx, stream, group, id).Err()
}

// Publish sends a message to a Redis pub/sub channel.
func (c *Client) Publish(ctx context.Context, channel string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return c.rdb.Publish(ctx, channel, string(data)).Err()
}

// Subscribe subscribes to a pub/sub channel and returns a channel of messages.
func (c *Client) Subscribe(ctx context.Context, channel string) (<-chan string, func()) {
	sub := c.rdb.Subscribe(ctx, channel)
	ch := make(chan string, 256)

	go func() {
		defer close(ch)
		for {
			msg, err := sub.ReceiveMessage(ctx)
			if err != nil {
				return
			}
			select {
			case ch <- msg.Payload:
			default:
				// drop if slow consumer
			}
		}
	}()

	return ch, func() { _ = sub.Close() }
}

// SetKey sets a key with optional TTL (0 = no expiry).
func (c *Client) SetKey(ctx context.Context, key string, v interface{}, ttl time.Duration) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, data, ttl).Err()
}

// GetKey gets a key and JSON-decodes into v.
func (c *Client) GetKey(ctx context.Context, key string, v interface{}) error {
	data, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// IncrKey increments a counter key.
func (c *Client) IncrKey(ctx context.Context, key string) (int64, error) {
	return c.rdb.Incr(ctx, key).Result()
}

func (c *Client) Close() error {
	return c.rdb.Close()
}
