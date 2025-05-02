package subpub

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestSubscribePublish(t *testing.T) {
    ps := NewSubPub()
    defer ps.Close(context.Background())

    var wg sync.WaitGroup
    wg.Add(1)

    received := make(chan bool, 1)

    _, err := ps.Subscribe("test", func(msg interface{}) {
        defer wg.Done()
        received <- true
    })
    if err != nil {
        t.Fatal(err)
    }

    err = ps.Publish("test", "message")
    if err != nil {
        t.Fatal(err)
    }

    wg.Wait()
    select {
    case <-received:
    default:
        t.Error("Message not received")
    }
}
func TestUnsubscribe(t *testing.T) {
	bus := NewSubPub()
	var callCount int

	sub, _ := bus.Subscribe("test", func(msg interface{}) {
		callCount++
	})

	bus.Publish("test", "msg1")
	time.Sleep(100 * time.Millisecond)
	sub.Unsubscribe()
	bus.Publish("test", "msg2")
	time.Sleep(100 * time.Millisecond)

	if callCount != 1 {
		t.Errorf("Expected handler to be called 1 time, got %d", callCount)
	}
}

func TestMultipleSubscribers(t *testing.T) {
	bus := NewSubPub()
	var wg sync.WaitGroup
	const subCount = 5

	wg.Add(subCount)
	for i := 0; i < subCount; i++ {
		_, _ = bus.Subscribe("test", func(msg interface{}) {
			wg.Done()
		})
	}

	bus.Publish("test", "msg")
	wg.Wait()
}

func TestSlowSubscriber(t *testing.T) {
	bus := NewSubPub()
	fastDone := make(chan struct{})
	slowDone := make(chan struct{})

	_, _ = bus.Subscribe("test", func(msg interface{}) {
		close(fastDone)
	})

	_, _ = bus.Subscribe("test", func(msg interface{}) {
		time.Sleep(500 * time.Millisecond)
		close(slowDone)
	})

	bus.Publish("test", "msg")

	select {
	case <-fastDone:

	case <-time.After(100 * time.Millisecond):
		t.Error("Fast subscriber blocked by slow one")
	}

	select {
	case <-slowDone:

	case <-time.After(600 * time.Millisecond):
		t.Error("Slow subscriber didn't complete")
	}
}

func TestClose(t *testing.T) {
	bus := NewSubPub()
	var handlerDone bool

	_, _ = bus.Subscribe("test", func(msg interface{}) {
		time.Sleep(200 * time.Millisecond)
		handlerDone = true
	})

	bus.Publish("test", "msg")
	
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	
	err := bus.Close(ctx)
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
	
	if !handlerDone {
		t.Error("Close didn't wait for handler to complete")
	}
}

func TestCloseWithCancel(t *testing.T) {
	bus := NewSubPub()
	var handlerDone bool

	_, _ = bus.Subscribe("test", func(msg interface{}) {
		time.Sleep(500 * time.Millisecond)
		handlerDone = true
	})

	bus.Publish("test", "msg")
	
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	
	err := bus.Close(ctx)
	if err != context.DeadlineExceeded {
		t.Errorf("Expected DeadlineExceeded, got %v", err)
	}
	
	if handlerDone {
		t.Error("Handler should still be running after canceled close")
	}
}

func TestPublishAfterClose(t *testing.T) {
	bus := NewSubPub()
	_, _ = bus.Subscribe("test", func(msg interface{}) {})

	err := bus.Close(context.Background())
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	err = bus.Publish("test", "msg")
	if err != context.Canceled {
		t.Errorf("Expected Canceled error, got %v", err)
	}
}

func TestSubscribeAfterClose(t *testing.T) {
	bus := NewSubPub()
	err := bus.Close(context.Background())
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	_, err = bus.Subscribe("test", func(msg interface{}) {})
	if err != context.Canceled {
		t.Errorf("Expected Canceled error, got %v", err)
	}
}
