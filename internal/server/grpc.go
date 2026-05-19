package server

import (
	"context"
	"fmt"
	"net"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GRPCServer struct {
	name     string
	server   *grpc.Server
	listener net.Listener
}

func NewGRPCServer(name string, host string, port int) (*GRPCServer, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("%s listen: %w", name, err)
	}

	gs := grpc.NewServer()

	srv := &GRPCServer{
		name:     name,
		server:   gs,
		listener: lis,
	}

	reflection.Register(gs)

	return srv, nil
}

func (s *GRPCServer) Server() *grpc.Server {
	return s.server
}

func (s *GRPCServer) Start() error {
	log.Info().Str("name", s.name).Str("addr", s.listener.Addr().String()).Msg("starting gRPC server")
	if err := s.server.Serve(s.listener); err != nil {
		return fmt.Errorf("%s serve: %w", s.name, err)
	}
	return nil
}

func (s *GRPCServer) Shutdown(ctx context.Context) error {
	log.Info().Str("name", s.name).Msg("shutting down gRPC server")
	done := make(chan struct{}, 1)
	go func() {
		s.server.GracefulStop()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
		s.server.Stop()
	}
	return nil
}

func (s *GRPCServer) Name() string {
	return s.name
}
