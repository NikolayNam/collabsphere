package realtime

import (
	"context"
	"sync"

	collabdomain "github.com/NikolayNam/collabsphere/internal/collab/domain"
	"github.com/google/uuid"
)

type Broker struct {
	mu   sync.RWMutex
	subs map[uuid.UUID]map[chan collabdomain.Event]struct{}
}

func NewBroker() *Broker {
	return &Broker{subs: make(map[uuid.UUID]map[chan collabdomain.Event]struct{})}
}

func (b *Broker) Publish(_ context.Context, event collabdomain.Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subs[event.ChannelID] {
		select {
		case ch <- event:
		default:
		}
	}
}

func (b *Broker) Subscribe(channelID uuid.UUID) (<-chan collabdomain.Event, func()) {
	ch := make(chan collabdomain.Event, 32)
	b.mu.Lock()
	if _, ok := b.subs[channelID]; !ok {
		b.subs[channelID] = make(map[chan collabdomain.Event]struct{})
	}
	b.subs[channelID][ch] = struct{}{}
	b.mu.Unlock()
	return ch, func() {
		b.mu.Lock()
		if subs, ok := b.subs[channelID]; ok {
			delete(subs, ch)
			if len(subs) == 0 {
				delete(b.subs, channelID)
			}
		}
		b.mu.Unlock()
		close(ch)
	}
}
