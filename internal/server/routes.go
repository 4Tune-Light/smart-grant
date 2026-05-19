package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"

	"github.com/rizky/smart-grant/internal/audit"
	"github.com/rizky/smart-grant/internal/auth"
	"github.com/rizky/smart-grant/internal/config"
	"github.com/rizky/smart-grant/internal/middleware"
	"github.com/rizky/smart-grant/internal/notification"
	"github.com/rizky/smart-grant/internal/proposal"
	"github.com/rizky/smart-grant/internal/review"
	"github.com/rizky/smart-grant/internal/risk"
	riskpb "github.com/rizky/smart-grant/proto/risk"
	notifpb "github.com/rizky/smart-grant/proto/notification"
	"github.com/rizky/smart-grant/pkg/storage"
)

type Services struct {
	Auth    *auth.Handler
	Proposal *proposal.Handler
	Review  *review.Handler
	Risk    *risk.Handler
	Audit   *audit.Handler
	Notif   *notification.Handler
}

func NewServices(cfg *config.Config, pool *pgxpool.Pool, rdb *redis.Client) *Services {
	return &Services{
		Auth:    newAuthHandler(cfg, pool),
		Proposal: newProposalHandler(cfg, pool, rdb),
		Review:  newReviewHandler(cfg, pool),
		Risk:    newRiskHandler(cfg, pool),
		Audit:   newAuditHandler(pool),
		Notif:   newNotifHandler(pool, rdb),
	}
}

func RegisterRoutes(r chi.Router, cfg *config.Config, pool *pgxpool.Pool, rdb *redis.Client) {
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)
	r.Use(middleware.CORS([]string{"*"}))
	r.Use(middleware.OTelHTTP(cfg.OTel.ServiceName + "-backend"))

	r.Get("/health", healthHandler)

	svc := NewServices(cfg, pool, rdb)

	registerAuthRoutes(r, cfg, svc.Auth)
	registerProposalRoutes(r, cfg, svc.Proposal)
	registerReviewRoutes(r, cfg, svc.Review)
	registerRiskRoutes(r, cfg, svc.Risk)
	registerAuditRoutes(r, cfg, svc.Audit)
	registerNotificationRoutes(r, cfg, svc.Notif)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func newAuthHandler(cfg *config.Config, pool *pgxpool.Pool) *auth.Handler {
	repo := auth.NewRepository(pool)
	svc := auth.NewService(repo, auth.TokenConfig{
		Secret:     cfg.JWT.Secret,
		AccessTTL:  cfg.JWT.AccessTTL,
		RefreshTTL: cfg.JWT.RefreshTTL,
	})
	return auth.NewHandler(svc)
}

func newProposalHandler(cfg *config.Config, pool *pgxpool.Pool, rdb *redis.Client) *proposal.Handler {
	proposalRepo := proposal.NewRepository(pool)
	auditSvc := audit.NewService(audit.NewRepository(pool))
	notifSvc := notification.NewService(notification.NewRepository(pool), rdb)

	minioStore, err := storage.NewMinio(storage.Config{
		Endpoint:  cfg.Storage.Minio.Endpoint,
		AccessKey: cfg.Storage.Minio.AccessKey,
		SecretKey: cfg.Storage.Minio.SecretKey,
		Bucket:    cfg.Storage.Minio.Bucket,
		UseSSL:    cfg.Storage.Minio.UseSSL,
		Region:    cfg.Storage.Minio.Region,
	})
	if err != nil {
		minioStore = nil
	}

	svc := proposal.NewService(proposalRepo, minioStore, auditSvc, notifSvc)
	return proposal.NewHandler(svc)
}

func newReviewHandler(cfg *config.Config, pool *pgxpool.Pool) *review.Handler {
	proposalRepo := proposal.NewRepository(pool)
	reviewRepo := review.NewRepository(pool)
	auditSvc := audit.NewService(audit.NewRepository(pool))
	notifSvc := notification.NewService(notification.NewRepository(pool), nil)

	svc := review.NewService(reviewRepo, proposalRepo, auditSvc, notifSvc)
	return review.NewHandler(svc)
}

func newRiskHandler(cfg *config.Config, pool *pgxpool.Pool) *risk.Handler {
	riskRepo := risk.NewRepository(pool)
	proposalRepo := proposal.NewRepository(pool)
	svc := risk.NewService(riskRepo, proposalRepo)
	return risk.NewHandler(svc)
}

func newAuditHandler(pool *pgxpool.Pool) *audit.Handler {
	repo := audit.NewRepository(pool)
	svc := audit.NewService(repo)
	return audit.NewHandler(svc)
}

func newNotifHandler(pool *pgxpool.Pool, rdb *redis.Client) *notification.Handler {
	repo := notification.NewRepository(pool)
	svc := notification.NewService(repo, rdb)
	return notification.NewHandler(svc)
}

func registerAuthRoutes(r chi.Router, cfg *config.Config, h *auth.Handler) {
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)
		r.Post("/refresh", h.RefreshToken)
	})

	r.Route("/api/v1/users", func(r chi.Router) {
		r.Use(middleware.Authenticate(cfg.JWT.Secret))
		r.Use(middleware.RequireRole("admin"))
		r.Get("/", h.ListUsers)
		r.Patch("/{id}/role", h.UpdateRole)
	})
}

func registerProposalRoutes(r chi.Router, cfg *config.Config, h *proposal.Handler) {
	r.Route("/api/v1/proposals", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(cfg.JWT.Secret))
			r.Post("/", h.Create)
			r.Get("/", h.List)
			r.Get("/page", h.ListPage)
			r.Get("/{id}", h.GetByID)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(cfg.JWT.Secret))
			r.Use(middleware.RequireRole("applicant"))
			r.Put("/{id}", h.Update)
			r.Post("/{id}/submit", h.Submit)
			r.Post("/{id}/documents", h.UploadDocument)
			r.Get("/{id}/documents", h.GetDocuments)
		})
	})
}

func registerReviewRoutes(r chi.Router, cfg *config.Config, h *review.Handler) {
	r.Route("/api/v1/reviews", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(cfg.JWT.Secret))
			r.Get("/{id}", h.GetByProposal)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(cfg.JWT.Secret))
			r.Use(middleware.RequireRole("reviewer"))
			r.Post("/{id}", h.Create)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(cfg.JWT.Secret))
			r.Use(middleware.RequireRole("admin"))
			r.Post("/{id}/approve", h.Approve)
			r.Post("/{id}/reject", h.Reject)
		})
	})
}

func registerRiskRoutes(r chi.Router, cfg *config.Config, h *risk.Handler) {
	r.Route("/api/v1/risk", func(r chi.Router) {
		r.Use(middleware.Authenticate(cfg.JWT.Secret))
		r.Use(middleware.RequireRole("admin", "reviewer"))
		r.Post("/{id}", h.Score)
		r.Get("/{id}", h.GetScore)
	})

	r.Route("/api/v1/risk/retrain", func(r chi.Router) {
		r.Use(middleware.Authenticate(cfg.JWT.Secret))
		r.Use(middleware.RequireRole("admin"))
		r.Post("/", h.Retrain)
	})
}

func registerAuditRoutes(r chi.Router, cfg *config.Config, h *audit.Handler) {
	r.Route("/api/v1/audit-logs", func(r chi.Router) {
		r.Use(middleware.Authenticate(cfg.JWT.Secret))
		r.Use(middleware.RequireRole("admin"))
		r.Get("/", h.List)
		r.Get("/{entity_id}", h.List)
	})
}

func registerNotificationRoutes(r chi.Router, cfg *config.Config, h *notification.Handler) {
	r.Route("/api/v1/notifications", func(r chi.Router) {
		r.Use(middleware.Authenticate(cfg.JWT.Secret))
		r.Get("/", h.List)
		r.Get("/stream", h.Stream)
		r.Patch("/read", h.MarkRead)
	})
}

func RegisterGRPC(s *grpc.Server, pool *pgxpool.Pool) {
	riskRepo := risk.NewRepository(pool)
	proposalRepo := proposal.NewRepository(pool)
	riskSvc := risk.NewService(riskRepo, proposalRepo)
	riskpb.RegisterRiskServiceServer(s, risk.NewGRPCServer(riskSvc))

	notifRepo := notification.NewRepository(pool)
	notifSvc := notification.NewService(notifRepo, nil)
	notifpb.RegisterNotificationServiceServer(s, notification.NewGRPCServer(notifSvc))
}
