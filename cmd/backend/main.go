package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/rizky/smart-grant/internal/auth"
	"github.com/rizky/smart-grant/internal/config"
	"github.com/rizky/smart-grant/internal/middleware"
	"github.com/rizky/smart-grant/internal/proposal"
	"github.com/rizky/smart-grant/internal/review"
	"github.com/rizky/smart-grant/internal/risk"
	"github.com/rizky/smart-grant/internal/server"
	"github.com/rizky/smart-grant/pkg/storage"
	"github.com/rizky/smart-grant/internal/telemetry"
	"github.com/rizky/smart-grant/pkg/database"
)

func main() {
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg := loadConfig()

	ctx := context.Background()

	tp, mp, err := telemetry.Init(ctx, telemetry.Config{
		ServiceName: cfg.OTel.ServiceName + "-backend",
		Environment: cfg.OTel.Environment,
		Endpoint:    cfg.OTel.Endpoint,
		Insecure:    cfg.OTel.Insecure,
		TraceRatio:  cfg.OTel.TraceRatio,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize telemetry")
	}
	defer func() {
		_ = tp.Shutdown(ctx)
		_ = mp.Shutdown(ctx)
	}()

	pgCfg := buildPostgresDSN(cfg)
	pgPool, err := database.NewPostgresPool(ctx, pgCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to PostgreSQL")
	}
	defer pgPool.Close()

	redisCfg := database.RedisConfig{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
	}
	if _, err := database.NewRedisClient(ctx, redisCfg); err != nil {
		log.Warn().Err(err).Msg("Redis not available, continuing without it")
	}

	httpSrv := server.NewHTTPServer(
		"backend-http",
		cfg.Server.HTTP.Host,
		cfg.Server.HTTP.Port,
		cfg.Server.HTTP.ReadTimeout,
	)

	registerRoutes(httpSrv.Router(), cfg, pgPool)

	grpcSrv, err := server.NewGRPCServer(
		"backend-grpc",
		cfg.Server.GRPC.Host,
		cfg.Server.GRPC.Port,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create gRPC server")
	}

	registerGRPCServices(grpcSrv.Server())

	mgr := server.NewManager(httpSrv, grpcSrv)

	log.Info().Msg("Backend service started")
	if err := mgr.Run(ctx); err != nil {
		log.Fatal().Err(err).Msg("server stopped with error")
	}
}

func loadConfig() *config.Config {
	cfgPath := os.Getenv("CONFIG_PATH")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}
	return cfg
}

func buildPostgresDSN(cfg *config.Config) database.PostgresConfig {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.Postgres.User,
		cfg.Database.Postgres.Password,
		cfg.Database.Postgres.Host,
		cfg.Database.Postgres.Port,
		cfg.Database.Postgres.DBName,
		cfg.Database.Postgres.SSLMode,
	)

	return database.PostgresConfig{
		DSN:             dsn,
		MaxOpenConns:    cfg.Database.Postgres.MaxOpenConns,
		MaxIdleConns:    cfg.Database.Postgres.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.Postgres.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.Database.Postgres.ConnMaxIdleTime,
	}
}

func registerRoutes(r chi.Router, cfg *config.Config, pool *pgxpool.Pool) {
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)
	r.Use(middleware.CORS([]string{"*"}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, auth.TokenConfig{
		Secret:     cfg.JWT.Secret,
		AccessTTL:  cfg.JWT.AccessTTL,
		RefreshTTL: cfg.JWT.RefreshTTL,
	})
	authHandler := auth.NewHandler(authSvc)

	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.RefreshToken)
	})

	proposalRepo := proposal.NewRepository(pool)

	minioStore, err := storage.NewMinio(storage.Config{
		Endpoint:  cfg.Storage.Minio.Endpoint,
		AccessKey: cfg.Storage.Minio.AccessKey,
		SecretKey: cfg.Storage.Minio.SecretKey,
		Bucket:    cfg.Storage.Minio.Bucket,
		UseSSL:    cfg.Storage.Minio.UseSSL,
		Region:    cfg.Storage.Minio.Region,
	})
	if err != nil {
		log.Warn().Err(err).Msg("MinIO not available, file upload will fail")
		minioStore = nil
	}

	proposalSvc := proposal.NewService(proposalRepo, minioStore)
	proposalHandler := proposal.NewHandler(proposalSvc)

	r.Route("/api/v1/proposals", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(cfg.JWT.Secret))
			r.Post("/", proposalHandler.Create)
			r.Get("/", proposalHandler.List)
			r.Get("/{id}", proposalHandler.GetByID)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(cfg.JWT.Secret))
			r.Use(middleware.RequireRole("applicant"))
			r.Put("/{id}", proposalHandler.Update)
			r.Post("/{id}/submit", proposalHandler.Submit)
			r.Post("/{id}/documents", proposalHandler.UploadDocument)
			r.Get("/{id}/documents", proposalHandler.GetDocuments)
		})
	})

	reviewRepo := review.NewRepository(pool)
	reviewSvc := review.NewService(reviewRepo, proposalRepo)
	reviewHandler := review.NewHandler(reviewSvc)

	r.Route("/api/v1/reviews", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(cfg.JWT.Secret))
			r.Get("/{id}", reviewHandler.GetByProposal)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(cfg.JWT.Secret))
			r.Use(middleware.RequireRole("reviewer"))
			r.Post("/{id}", reviewHandler.Create)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(cfg.JWT.Secret))
			r.Use(middleware.RequireRole("admin"))
			r.Post("/{id}/approve", reviewHandler.Approve)
			r.Post("/{id}/reject", reviewHandler.Reject)
		})
	})

	riskRepo := risk.NewRepository(pool)
	riskSvc := risk.NewService(riskRepo, proposalRepo)
	riskHandler := risk.NewHandler(riskSvc)

	r.Route("/api/v1/risk", func(r chi.Router) {
		r.Use(middleware.Authenticate(cfg.JWT.Secret))
		r.Use(middleware.RequireRole("admin", "reviewer"))
		r.Post("/{id}", riskHandler.Score)
		r.Get("/{id}", riskHandler.GetScore)
	})
}

func registerGRPCServices(s *grpc.Server) {
}
