package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

type Server interface {
	Start() error
	Shutdown(ctx context.Context) error
	Name() string
}

type Manager struct {
	servers []Server
}

func NewManager(servers ...Server) *Manager {
	return &Manager{
		servers: servers,
	}
}

func (m *Manager) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	errCh := make(chan error, len(m.servers))

	for _, srv := range m.servers {
		srv := srv
		go func() {
			if err := srv.Start(); err != nil {
				errCh <- err
			}
		}()
	}

	select {
	case <-ctx.Done():
		return m.gracefulShutdown()
	case err := <-errCh:
		return err
	}
}

func (m *Manager) gracefulShutdown() error {
	for _, srv := range m.servers {
		if err := srv.Shutdown(context.Background()); err != nil {
			return err
		}
	}
	return nil
}
