package notification

import (
	"context"
	"fmt"

	pb "github.com/rizky/smart-grant/proto/notification"
)

type gRPCServer struct {
	pb.UnimplementedNotificationServiceServer
	svc Service
}

func NewGRPCServer(svc Service) pb.NotificationServiceServer {
	return &gRPCServer{svc: svc}
}

func (s *gRPCServer) SendNotification(ctx context.Context, req *pb.SendNotificationRequest) (*pb.SendNotificationResponse, error) {
	err := s.svc.Send(ctx, req.UserId, req.Type, req.Title, req.Body)
	if err != nil {
		return nil, fmt.Errorf("send notification: %w", err)
	}

	return &pb.SendNotificationResponse{
		Success: true,
	}, nil
}
