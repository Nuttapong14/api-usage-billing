package redis

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrPubSubClosed    = errors.New("pubsub connection closed")
	ErrSubscribeFailed = errors.New("subscribe failed")
)

// Message represents a pub/sub message
type Message struct {
	Channel   string
	Payload   string
	Pattern   string
	Timestamp time.Time
}

// MessageHandler handles incoming pub/sub messages
type MessageHandler func(ctx context.Context, msg *Message) error

// PubSub provides Redis pub/sub functionality
type PubSub struct {
	client       *Client
	subscriptions map[string]*redis.PubSub
	handlers     map[string][]MessageHandler
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// NewPubSub creates a new pub/sub instance
func NewPubSub(client *Client) *PubSub {
	ctx, cancel := context.WithCancel(context.Background())
	return &PubSub{
		client:        client,
		subscriptions: make(map[string]*redis.PubSub),
		handlers:      make(map[string][]MessageHandler),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Publish publishes a message to a channel
func (ps *PubSub) Publish(ctx context.Context, channel string, payload interface{}) error {
	var data string
	switch v := payload.(type) {
	case string:
		data = v
	case []byte:
		data = string(v)
	default:
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		data = string(jsonData)
	}

	return ps.client.rdb.Publish(ctx, ps.client.prefixKey(channel), data).Err()
}

// Subscribe subscribes to channels
func (ps *PubSub) Subscribe(ctx context.Context, handler MessageHandler, channels ...string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	prefixedChannels := make([]string, len(channels))
	for i, ch := range channels {
		prefixedChannels[i] = ps.client.prefixKey(ch)
	}

	pubsub := ps.client.rdb.Subscribe(ctx, prefixedChannels...)

	// Wait for subscription confirmation
	_, err := pubsub.Receive(ctx)
	if err != nil {
		pubsub.Close()
		return err
	}

	// Store subscription and handlers
	for _, ch := range channels {
		ps.subscriptions[ch] = pubsub
		ps.handlers[ch] = append(ps.handlers[ch], handler)
	}

	// Start message receiver
	ps.wg.Add(1)
	go ps.receiveMessages(pubsub, channels)

	return nil
}

// PSubscribe subscribes to channels matching patterns
func (ps *PubSub) PSubscribe(ctx context.Context, handler MessageHandler, patterns ...string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	prefixedPatterns := make([]string, len(patterns))
	for i, p := range patterns {
		prefixedPatterns[i] = ps.client.prefixKey(p)
	}

	pubsub := ps.client.rdb.PSubscribe(ctx, prefixedPatterns...)

	// Wait for subscription confirmation
	_, err := pubsub.Receive(ctx)
	if err != nil {
		pubsub.Close()
		return err
	}

	// Store subscription and handlers
	for _, p := range patterns {
		ps.subscriptions[p] = pubsub
		ps.handlers[p] = append(ps.handlers[p], handler)
	}

	// Start message receiver
	ps.wg.Add(1)
	go ps.receivePatternMessages(pubsub, patterns)

	return nil
}

// receiveMessages receives messages from subscribed channels
func (ps *PubSub) receiveMessages(pubsub *redis.PubSub, channels []string) {
	defer ps.wg.Done()

	ch := pubsub.Channel()

	for {
		select {
		case <-ps.ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}

			ps.handleMessage(&Message{
				Channel:   msg.Channel,
				Payload:   msg.Payload,
				Timestamp: time.Now(),
			}, channels)
		}
	}
}

// receivePatternMessages receives messages from pattern subscriptions
func (ps *PubSub) receivePatternMessages(pubsub *redis.PubSub, patterns []string) {
	defer ps.wg.Done()

	ch := pubsub.Channel()

	for {
		select {
		case <-ps.ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}

			ps.handleMessage(&Message{
				Channel:   msg.Channel,
				Payload:   msg.Payload,
				Pattern:   msg.Pattern,
				Timestamp: time.Now(),
			}, patterns)
		}
	}
}

// handleMessage dispatches a message to registered handlers
func (ps *PubSub) handleMessage(msg *Message, keys []string) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	for _, key := range keys {
		handlers, ok := ps.handlers[key]
		if !ok {
			continue
		}

		for _, handler := range handlers {
			// Create a context with timeout for handler
			ctx, cancel := context.WithTimeout(ps.ctx, 30*time.Second)
			if err := handler(ctx, msg); err != nil {
				// Log error but continue processing
				// In production, this should use a proper logger
			}
			cancel()
		}
	}
}

// Unsubscribe unsubscribes from channels
func (ps *PubSub) Unsubscribe(ctx context.Context, channels ...string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	for _, ch := range channels {
		if pubsub, ok := ps.subscriptions[ch]; ok {
			if err := pubsub.Unsubscribe(ctx, ps.client.prefixKey(ch)); err != nil {
				return err
			}
			delete(ps.subscriptions, ch)
			delete(ps.handlers, ch)
		}
	}

	return nil
}

// PUnsubscribe unsubscribes from patterns
func (ps *PubSub) PUnsubscribe(ctx context.Context, patterns ...string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	for _, p := range patterns {
		if pubsub, ok := ps.subscriptions[p]; ok {
			if err := pubsub.PUnsubscribe(ctx, ps.client.prefixKey(p)); err != nil {
				return err
			}
			delete(ps.subscriptions, p)
			delete(ps.handlers, p)
		}
	}

	return nil
}

// Close closes all pub/sub connections
func (ps *PubSub) Close() error {
	ps.cancel()

	ps.mu.Lock()
	for _, pubsub := range ps.subscriptions {
		pubsub.Close()
	}
	ps.subscriptions = make(map[string]*redis.PubSub)
	ps.handlers = make(map[string][]MessageHandler)
	ps.mu.Unlock()

	ps.wg.Wait()
	return nil
}

// ChannelCount returns the number of active subscriptions
func (ps *PubSub) ChannelCount() int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return len(ps.subscriptions)
}
