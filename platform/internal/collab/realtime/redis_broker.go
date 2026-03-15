package realtime

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	collabdomain "github.com/NikolayNam/collabsphere/internal/collab/domain"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// RedisBroker publishes and subscribes to realtime events via Redis Pub/Sub.
// Events are split by channel: each collab channel has its own Redis channel.
// On init, loads collab.channels and collab.messages from Postgres; new subscribers get cached history first.
// When Redis is unavailable, events are stored in EventBufferStore (Postgres) and drained when Redis recovers.
type RedisBroker struct {
	client        *redis.Client
	channelPrefix string
	buffer        EventBufferStore
	loader        ChannelLoader
	initLimit     int
	cacheMu       sync.RWMutex
	cache         map[uuid.UUID][]collabdomain.Event
	drainInterval time.Duration
	drainLimit    int
	stopCh        chan struct{}
	stopOnce      sync.Once
}

// RedisBrokerOption configures RedisBroker.
type RedisBrokerOption func(*RedisBroker)

// WithEventBuffer enables Postgres fallback when Redis publish fails.
func WithEventBuffer(store EventBufferStore) RedisBrokerOption {
	return func(b *RedisBroker) {
		b.buffer = store
	}
}

// WithChannelLoader loads collab.channels and collab.messages on init for subscriber backfill.
func WithChannelLoader(loader ChannelLoader) RedisBrokerOption {
	return func(b *RedisBroker) {
		b.loader = loader
	}
}

// WithInitMessageLimit sets max messages per channel to load on init. Default 100.
func WithInitMessageLimit(limit int) RedisBrokerOption {
	return func(b *RedisBroker) {
		b.initLimit = limit
	}
}

// WithDrainInterval sets how often to attempt draining the buffer to Redis. Default 5s.
func WithDrainInterval(d time.Duration) RedisBrokerOption {
	return func(b *RedisBroker) {
		b.drainInterval = d
	}
}

// NewRedisBroker creates a broker that uses Redis Pub/Sub.
func NewRedisBroker(client *redis.Client, channelPrefix string, opts ...RedisBrokerOption) *RedisBroker {
	b := &RedisBroker{
		client:        client,
		channelPrefix: channelPrefix,
		cache:         make(map[uuid.UUID][]collabdomain.Event),
		drainInterval: 5 * time.Second,
		drainLimit:    100,
		initLimit:     100,
		stopCh:        make(chan struct{}),
	}
	for _, opt := range opts {
		opt(b)
	}
	if b.loader != nil {
		go b.initLoad()
	}
	if b.buffer != nil {
		go b.drainLoop()
	}
	return b
}

func (b *RedisBroker) initLoad() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	channelIDs, err := b.loader.ListAllChannelIDs(ctx)
	if err != nil {
		slog.Warn("redis broker: init load channels failed", "err", err)
		return
	}
	limit := b.initLimit
	if limit <= 0 {
		limit = 100
	}
	for _, channelID := range channelIDs {
		msgs, err := b.loader.ListRecentMessagesForChannel(ctx, channelID, limit)
		if err != nil {
			slog.Warn("redis broker: init load messages failed", "channelId", channelID, "err", err)
			continue
		}
		events := make([]collabdomain.Event, 0, len(msgs))
		for _, m := range msgs {
			events = append(events, collabdomain.Event{Type: "message.created", ChannelID: channelID, Payload: m})
		}
		if len(events) > 0 {
			b.cacheMu.Lock()
			b.cache[channelID] = events
			b.cacheMu.Unlock()
		}
	}
	slog.Info("redis broker: init load complete", "channels", len(channelIDs))
}

// Stop stops the drain goroutine. Call on shutdown.
func (b *RedisBroker) Stop() {
	b.stopOnce.Do(func() { close(b.stopCh) })
}

func (b *RedisBroker) channelKey(channelID uuid.UUID) string {
	return b.channelPrefix + ":channel:" + channelID.String()
}

// Publish serializes the event to JSON and publishes it to the Redis channel for the event's ChannelID.
// Also appends message.created events to the per-channel cache for new subscribers.
// If Redis fails and EventBufferStore is configured, the event is stored in Postgres for later drain.
func (b *RedisBroker) Publish(ctx context.Context, event collabdomain.Event) {
	payload, err := json.Marshal(event)
	if err != nil {
		slog.Error("redis broker: marshal event", "err", err, "channelId", event.ChannelID)
		return
	}
	if event.Type == "message.created" {
		b.appendToCache(event)
	}
	if err := b.client.Publish(ctx, b.channelKey(event.ChannelID), payload).Err(); err != nil {
		slog.Warn("redis broker: publish failed, storing in buffer", "err", err, "channelId", event.ChannelID)
		if b.buffer != nil {
			if storeErr := b.buffer.Store(ctx, event); storeErr != nil {
				slog.Error("redis broker: buffer store failed", "err", storeErr, "channelId", event.ChannelID)
			}
		}
	}
}

func (b *RedisBroker) appendToCache(event collabdomain.Event) {
	const maxCached = 200
	b.cacheMu.Lock()
	defer b.cacheMu.Unlock()
	slice := b.cache[event.ChannelID]
	if len(slice) >= maxCached {
		slice = append(slice[1:], event)
	} else {
		slice = append(slice, event)
	}
	b.cache[event.ChannelID] = slice
}

// Subscribe subscribes to the Redis channel for the given channelID and returns a local channel
// that receives events. Sends cached init events first (from collab.messages), then streams from Redis.
// The returned function unsubscribes and closes the channel.
func (b *RedisBroker) Subscribe(channelID uuid.UUID) (<-chan collabdomain.Event, func()) {
	ch := make(chan collabdomain.Event, 64)
	ctx, cancel := context.WithCancel(context.Background())
	pubsub := b.client.Subscribe(ctx, b.channelKey(channelID))

	go func() {
		defer close(ch)
		defer pubsub.Close()
		b.cacheMu.RLock()
		cached := b.cache[channelID]
		b.cacheMu.RUnlock()
		for _, ev := range cached {
				select {
				case <-ctx.Done():
					return
				case ch <- ev:
				default:
					// channel full, skip rest of cache
					goto doneCache
				}
			}
		doneCache:
		for {
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil {
				return
			}
			var event collabdomain.Event
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				slog.Warn("redis broker: unmarshal event", "err", err, "channelId", channelID)
				continue
			}
			select {
			case ch <- event:
			default:
				// channel full, drop
			}
		}
	}()

	unsubscribe := func() {
		cancel()
	}
	return ch, unsubscribe
}

func (b *RedisBroker) drainLoop() {
	ticker := time.NewTicker(b.drainInterval)
	defer ticker.Stop()
	for {
		select {
		case <-b.stopCh:
			return
		case <-ticker.C:
			b.drain(context.Background())
		}
	}
}

func (b *RedisBroker) drain(ctx context.Context) {
	if b.buffer == nil {
		return
	}
	events, ids, err := b.buffer.Drain(ctx, b.drainLimit)
	if err != nil {
		slog.Warn("redis broker: drain fetch failed", "err", err)
		return
	}
	if len(events) == 0 {
		return
	}
	var published []uuid.UUID
	for i, event := range events {
		payload, err := json.Marshal(event)
		if err != nil {
			continue
		}
		if err := b.client.Publish(ctx, b.channelKey(event.ChannelID), payload).Err(); err != nil {
			slog.Warn("redis broker: drain publish failed", "err", err, "channelId", event.ChannelID)
			break
		}
		published = append(published, ids[i])
	}
	if len(published) > 0 {
		if err := b.buffer.Delete(ctx, published); err != nil {
			slog.Warn("redis broker: drain delete failed", "err", err)
		}
	}
}
