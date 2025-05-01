package subpub

import (
	"context"
	"sync"
)

// MessageHandler is a callback function that processes messages delivered to subscribers
type MessageHandler func(msg interface{})

type Subscription interface {
	// Unsubscribe will remove interest in the current subject subscription
	Unsubscribe()
}

type SubPub interface {
	Subscribe(subject string, cb MessageHandler) (Subscription, error)
	Publish(subject string, msg interface{}) error
	Close(ctx context.Context) error
}

type subscription struct {
	subject  string
	handler  MessageHandler
	messages chan interface{}
	bus      *subPubImpl
}

func (s *subscription) Unsubscribe() {
	s.bus.unsubscribe(s)
}

type subPubImpl struct {
	mu          sync.RWMutex
	subscribers map[string][]*subscription
	closed      bool
	wg          sync.WaitGroup
}

func NewSubPub() SubPub {
	return &subPubImpl{
		subscribers: make(map[string][]*subscription),
	}
}

func (b *subPubImpl) Subscribe(subject string, cb MessageHandler) (Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil, context.Canceled
	}

	sub := &subscription{
		subject:  subject,
		handler:  cb,
		messages: make(chan interface{}, 100), // Buffered channel to prevent blocking publishers
		bus:      b,
	}

	b.subscribers[subject] = append(b.subscribers[subject], sub)
	b.wg.Add(1)

	go func() {
		defer b.wg.Done()
		for msg := range sub.messages {
			sub.handler(msg)
		}
	}()

	return sub, nil
}

func (b *subPubImpl) Publish(subject string, msg interface{}) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return context.Canceled
	}

	subs, ok := b.subscribers[subject]
	if !ok {
		return nil
	}

	for _, sub := range subs {
		select {
		case sub.messages <- msg:
		default:
			// Skip if subscriber's channel is full to prevent blocking
		}
	}

	return nil
}

func (b *subPubImpl) unsubscribe(sub *subscription) {
	b.mu.Lock()
	defer b.mu.Unlock()

	subs, ok := b.subscribers[sub.subject]
	if !ok {
		return
	}

	for i, s := range subs {
		if s == sub {
			// Remove the subscription
			b.subscribers[sub.subject] = append(subs[:i], subs[i+1:]...)
			close(sub.messages)
			break
		}
	}

	if len(b.subscribers[sub.subject]) == 0 {
		delete(b.subscribers, sub.subject)
	}
}

func (b *subPubImpl) Close(ctx context.Context) error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil
	}
	b.closed = true

	// Collect all subscriptions
	var allSubs []*subscription
	for _, subs := range b.subscribers {
		allSubs = append(allSubs, subs...)
	}
	b.subscribers = nil
	b.mu.Unlock()

	// Close all subscription channels
	for _, sub := range allSubs {
		close(sub.messages)
	}

	// Wait for all handlers to finish or context to be canceled
	done := make(chan struct{})
	go func() {
		b.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}