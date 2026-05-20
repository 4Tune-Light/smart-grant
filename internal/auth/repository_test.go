package auth

import (
	"context"
	"testing"

	"github.com/rizky/smart-grant/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestIntegration_AuthRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image: "postgres:16-alpine",
		Env: map[string]string{
			"POSTGRES_USER": "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":   "testdb",
		},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForLog("database system is ready to accept connections"),
	}

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer pgContainer.Terminate(ctx)

	host, _ := pgContainer.Host(ctx)
	port, _ := pgContainer.MappedPort(ctx, "5432")
	dsn := "postgres://testuser:testpass@" + host + ":" + port.Port() + "/testdb?sslmode=disable"

	pool := newTestPool(t, ctx, dsn)
	defer pool.Close()

	runMigration(t, pool)

	q := database.NewQuerier(pool)
	repo := NewRepository(q)

	user := &User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "applicant"}
	err = repo.Create(ctx, user)
	require.NoError(t, err)
	assert.NotEmpty(t, user.ID)

	found, err := repo.FindByEmail(ctx, "test@example.com")
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", found.Email)

	foundByID, err := repo.FindByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, foundByID.ID)

	dup := &User{Email: "test@example.com", PasswordHash: "hash2", Name: "Dup", Role: "applicant"}
	err = repo.Create(ctx, dup)
	assert.ErrorIs(t, err, ErrEmailAlreadyExists)
}
