package risk

import (
	"context"
	"fmt"

	pb "github.com/rizky/smart-grant/proto/risk"
)

type gRPCServer struct {
	pb.UnimplementedRiskServiceServer
	svc Service
}

func NewGRPCServer(svc Service) pb.RiskServiceServer {
	return &gRPCServer{svc: svc}
}

func (s *gRPCServer) CalculateRisk(ctx context.Context, req *pb.CalculateRiskRequest) (*pb.CalculateRiskResponse, error) {
	resp, err := s.svc.Score(ctx, req.ProposalId)
	if err != nil {
		return nil, fmt.Errorf("calculate risk: %w", err)
	}

	features := make(map[string]float64)
	for k, v := range resp.Features {
		features[k] = v
	}

	return &pb.CalculateRiskResponse{
		ProposalId:        resp.ProposalID,
		RiskLevel:         resp.RiskLevel,
		Confidence:        resp.Confidence,
		FeatureImportance: features,
	}, nil
}
