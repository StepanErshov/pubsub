package service

import (
	"context"
	"testing"

	"github.com/StepanErshov/pubsub/pkg/pb"
	"github.com/StepanErshov/pubsub/pkg/subpub"
	"google.golang.org/grpc"
)

func TestPubSubService(t *testing.T) {
	bus := subpub.NewSubPub()
	service := NewPubSubService(bus)

	server := grpc.NewServer()
	service.Register(server)
}

func TestPublish(t *testing.T) {
	bus := subpub.NewSubPub()
	service := NewPubSubService(bus)

	_, err := service.Publish(context.Background(), &pb.PublishRequest{
		Key:  "test",
		Data: "message",
	})
	if err != nil {
		t.Errorf("Publish failed: %v", err)
	}
}