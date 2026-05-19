package risk

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/rizky/smart-grant/internal/middleware"
	"github.com/rizky/smart-grant/internal/proposal"
	"github.com/rizky/smart-grant/internal/risk/engine"
)

type Service interface {
	Score(ctx context.Context, proposalID string) (*RiskResponse, error)
	GetScore(ctx context.Context, proposalID string) (*RiskResponse, error)
}

type service struct {
	repo         Repository
	proposalRepo proposal.Repository
	tree         *engine.DecisionTree
	once         sync.Once
}

func NewService(repo Repository, proposalRepo proposal.Repository) Service {
	return &service{
		repo:         repo,
		proposalRepo: proposalRepo,
	}
}

func (s *service) ensureTree() {
	s.once.Do(func() {
		data := engine.GenerateSeedData()
		s.tree = engine.BuildTree(data, engine.RiskFeatures)
	})
}

func (s *service) Score(ctx context.Context, proposalID string) (*RiskResponse, error) {
	ctx, span := otel.Tracer("smart-grant").Start(ctx, "risk.Score")
	defer span.End()

	role, _ := ctx.Value(middleware.AuthRoleKey).(string)
	if role != "admin" && role != "reviewer" {
		return nil, fmt.Errorf("insufficient permissions")
	}

	prop, err := s.proposalRepo.FindByID(ctx, proposalID)
	if err != nil {
		return nil, ErrProposalNotFound
	}

	span.SetAttributes(
		attribute.String("proposal.id", proposalID),
		attribute.Float64("proposal.nominal_amount", prop.NominalAmount),
		attribute.String("proposal.organization", prop.Organization),
	)

	s.ensureTree()

	features := extractFeatures(prop)
	label, confidence := s.tree.Classify(features)

	span.SetAttributes(
		attribute.String("risk.level", string(label)),
		attribute.Float64("risk.confidence", float64(confidence)),
	)

	featuresJSON, _ := json.Marshal(features)
	details := fmt.Sprintf(`{"tree_depth": %d}`, treeDepth(s.tree.Root))

	score := &RiskScore{
		ProposalID:   proposalID,
		RiskLevel:    string(label),
		Confidence:   float64(confidence),
		Features:     string(featuresJSON),
		Details:      details,
		ModelVersion: "c4.5-v1",
	}

	if err := s.repo.Save(ctx, score); err != nil {
		return nil, err
	}

	return toResponse(score, features), nil
}

func (s *service) GetScore(ctx context.Context, proposalID string) (*RiskResponse, error) {
	score, err := s.repo.FindByProposalID(ctx, proposalID)
	if err != nil {
		return nil, err
	}
	if score == nil {
		return nil, ErrScoreNotFound
	}

	var features map[string]float64
	json.Unmarshal([]byte(score.Features), &features)

	return toResponse(score, features), nil
}

func extractFeatures(prop *proposal.Proposal) map[string]float64 {
	freq := 0.0
	completeness := 1.0

	return map[string]float64{
		"nominal_amount":         prop.NominalAmount,
		"funding_frequency_30d":  freq,
		"document_completeness":  completeness,
	}
}

func treeDepth(node *engine.TreeNode) int {
	if node.IsLeaf {
		return 1
	}
	leftDepth := treeDepth(node.LessEqual)
	rightDepth := treeDepth(node.GreaterThan)
	if leftDepth > rightDepth {
		return leftDepth + 1
	}
	return rightDepth + 1
}

func toResponse(score *RiskScore, features map[string]float64) *RiskResponse {
	return &RiskResponse{
		ID:           score.ID,
		ProposalID:   score.ProposalID,
		RiskLevel:    score.RiskLevel,
		Confidence:   score.Confidence,
		Features:     features,
		ModelVersion: score.ModelVersion,
		CreatedAt:    score.CreatedAt.Format(time.RFC3339),
	}
}
