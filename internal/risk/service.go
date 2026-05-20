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
	riskdto "github.com/rizky/smart-grant/internal/risk/dto"
	"github.com/rizky/smart-grant/internal/risk/engine"
)

type Service interface {
	Score(ctx context.Context, proposalID string) (*riskdto.RiskResponse, error)
	GetScore(ctx context.Context, proposalID string) (*riskdto.RiskResponse, error)
	Retrain(ctx context.Context) (*RetrainResponse, error)
}

type RetrainResponse struct {
	PreviousVersion string `json:"previous_version"`
	NewVersion      string `json:"new_version"`
	ExamplesUsed    int    `json:"examples_used"`
	TreeDepth       int    `json:"tree_depth"`
}

type service struct {
	repo         Repository
	proposalRepo proposal.Repository
	tree         *engine.DecisionTree
	version      string
	mu           sync.RWMutex
	once         sync.Once
}

func NewService(repo Repository, proposalRepo proposal.Repository) Service {
	return &service{
		repo:         repo,
		proposalRepo: proposalRepo,
		version:      engine.ModelVersion,
	}
}

func (s *service) ensureTree() {
	s.once.Do(func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.tree == nil {
			data := engine.GenerateSeedData()
			s.tree = engine.BuildTree(data, engine.RiskFeatures)
		}
	})
}

func (s *service) getTree() *engine.DecisionTree {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tree
}

func (s *service) Score(ctx context.Context, proposalID string) (*riskdto.RiskResponse, error) {
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

	tree := s.getTree()
	features := s.extractFeatures(ctx, prop)
	label, confidence := tree.Classify(features)

	s.mu.RLock()
	version := s.version
	s.mu.RUnlock()

	span.SetAttributes(
		attribute.String("risk.level", string(label)),
		attribute.Float64("risk.confidence", float64(confidence)),
	)

	featuresJSON, _ := json.Marshal(features)
	details := fmt.Sprintf(`{"tree_depth": %d}`, treeDepth(tree.Root))

	score := &RiskScore{
		ProposalID:   proposalID,
		RiskLevel:    string(label),
		Confidence:   float64(confidence),
		Features:     string(featuresJSON),
		Details:      details,
		ModelVersion: version,
	}

	if err := s.repo.Save(ctx, score); err != nil {
		return nil, err
	}

	return toResponse(score, features), nil
}

func (s *service) GetScore(ctx context.Context, proposalID string) (*riskdto.RiskResponse, error) {
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

func (s *service) Retrain(ctx context.Context) (*RetrainResponse, error) {
	s.mu.RLock()
	oldVersion := s.version
	s.mu.RUnlock()

	scores, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("query risk scores: %w", err)
	}

	if len(scores) < 5 {
		return nil, fmt.Errorf("need at least 5 scored proposals to retrain, got %d", len(scores))
	}

	examples := make([]engine.Example, 0, len(scores))
	for _, sc := range scores {
		var features map[string]float64
		if err := json.Unmarshal([]byte(sc.Features), &features); err != nil {
			continue
		}
		examples = append(examples, engine.Example{
			Features: features,
			Label:    engine.Label(sc.RiskLevel),
		})
	}

	newTree := engine.BuildTree(examples, engine.RiskFeatures)
	newVersion := fmt.Sprintf("c4.5-v2-retrain-%d", len(scores))

	s.mu.Lock()
	s.tree = newTree
	s.version = newVersion
	s.mu.Unlock()

	return &RetrainResponse{
		PreviousVersion: oldVersion,
		NewVersion:      newVersion,
		ExamplesUsed:    len(examples),
		TreeDepth:       treeDepth(newTree.Root),
	}, nil
}

func (s *service) extractFeatures(ctx context.Context, prop *proposal.Proposal) map[string]float64 {
	since := time.Now().AddDate(0, 0, -30)
	freq, _ := s.proposalRepo.CountByOrganization(ctx, prop.Organization, since)

	docCount, _ := s.proposalRepo.CountDocuments(ctx, prop.ID)
	completeness := 1.0
	if docCount == 0 {
		completeness = 0.0
	}

	return map[string]float64{
		"nominal_amount":         prop.NominalAmount,
		"funding_frequency_30d":  float64(freq),
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

func toResponse(score *RiskScore, features map[string]float64) *riskdto.RiskResponse {
	return &riskdto.RiskResponse{
		ID:           score.ID,
		ProposalID:   score.ProposalID,
		RiskLevel:    score.RiskLevel,
		Confidence:   score.Confidence,
		Features:     features,
		ModelVersion: score.ModelVersion,
		CreatedAt:    score.CreatedAt.Format(time.RFC3339),
	}
}
