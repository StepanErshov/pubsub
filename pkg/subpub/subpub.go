package subpub

import (
	"context"
	"sync"
)

type MessageHandler func(msg interface{})

type Subscription interface {
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
	once     sync.Once
}

func (s *subscription) Unsubscribe() {
	s.bus.unsubscribe(s)
}

type subPubImpl struct {
	mu          sync.RWMutex
	subscribers map[string][]*subscription
	closed      bool
	wg          sync.WaitGroup
	closeOnce   sync.Once
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
		messages: make(chan interface{}, 100),
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
		}
	}

	return nil
}

func (b *subPubImpl) unsubscribe(sub *subscription) {
	sub.once.Do(func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		subs, ok := b.subscribers[sub.subject]
		if !ok {
			return
		}

		for i, s := range subs {
			if s == sub {
				b.subscribers[sub.subject] = append(subs[:i], subs[i+1:]...)
				close(sub.messages)
				break
			}
		}

		if len(b.subscribers[sub.subject]) == 0 {
			delete(b.subscribers, sub.subject)
		}
	})
}

func (b *subPubImpl) Close(ctx context.Context) error {
	var err error
	b.closeOnce.Do(func() {
		b.mu.Lock()
		b.closed = true
		subsCopy := make(map[string][]*subscription)
		for k, v := range b.subscribers {
			subs := make([]*subscription, len(v))
			copy(subs, v)
			subsCopy[k] = subs
		}
		b.subscribers = nil
		b.mu.Unlock()

		for _, subs := range subsCopy {
			for _, sub := range subs {
				sub.once.Do(func() {
					close(sub.messages)
				})
			}
		}

		done := make(chan struct{})
		go func() {
			b.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			err = nil
		case <-ctx.Done():
			err = ctx.Err()
		}
	})
	return err
}