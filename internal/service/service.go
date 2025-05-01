package service

import (
	"context"
    "github.com/rs/zerolog/log"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/protobuf/types/known/emptypb"
    
    "github.com/StepanErshov/pubsub/pkg/pb"
    "github.com/StepanErshov/pubsub/pkg/subpub"
)

type PubSubService struct {
	pb.UnimplementedPubSubServer
	bus subpub.SubPub
}

func NewPubSubService(bus subpub.SubPub) *PubSubService {
	return &PubSubService{bus: bus}
}

func (s *PubSubService) Register(server *grpc.Server) {
    pb.RegisterPubSubServer(server, s)
    log.Info().Msg("PubSub service registered")
}

func (s *PubSubService) Subscribe(req *pb.SubscribeRequest, stream pb.PubSub_SubscribeServer) error {
	key := req.GetKey()
	if key == "" {
		return status.Error(codes.InvalidArgument, "key is required")
	}

	log.Info().Str("key", key).Msg("New subscription")

	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	sub, err := s.bus.Subscribe(key, func(msg interface{}) {
		data, ok := msg.(string)
		if !ok {
			log.Error().Msg("Invalid message type")
			return
		}

		if err := stream.Send(&pb.Event{Data: data}); err != nil {
			log.Error().Err(err).Msg("Failed to send event")
			cancel()
		}
	})
	if err != nil {
		return status.Error(codes.Internal, "failed to subscribe")
	}
	defer sub.Unsubscribe()

	<-ctx.Done()
	return nil
}

func (s *PubSubService) Publish(ctx context.Context, req *pb.PublishRequest) (*emptypb.Empty, error) {
	key := req.GetKey()
	data := req.GetData()

	if key == "" {
		return nil, status.Error(codes.InvalidArgument, "key is required")
	}

	if err := s.bus.Publish(key, data); err != nil {
		return nil, status.Error(codes.Internal, "failed to publish")
	}

	log.Info().Str("key", key).Str("data", data).Msg("Published event")
	return &emptypb.Empty{}, nil
}