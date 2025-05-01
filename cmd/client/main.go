package main

import (
	"context"
	"log"
	"time"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/StepanErshov/pubsub/pkg/pb"
)

func main() {
	conn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewPubSubClient(conn)

	go func() {
		stream, err := client.Subscribe(context.Background(), &pb.SubscribeRequest{Key: "test"})
		if err != nil {
			log.Fatal(err)
		}
		for {
			event, err := stream.Recv()
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Received: %s", event.Data)
		}
	}()

	for i := 0; i < 3; i++ {
		_, err = client.Publish(context.Background(), &pb.PublishRequest{
			Key:  "test",
			Data: fmt.Sprintf("Message %d", i),
		})
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(1 * time.Second)
	}
}