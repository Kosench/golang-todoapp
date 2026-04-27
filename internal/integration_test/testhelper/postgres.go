package testhelper

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"time"

	core_postgres_pool "github.com/Kosench/golang-todoapp/internal/core/repository/postgres/pool"
	core_pgx_pool "github.com/Kosench/golang-todoapp/internal/core/repository/postgres/pool/pgx"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer хранит запущенный контейнер и пул соединений.
type PostgresContainer struct {
	Container testcontainers.Container
	Pool      core_postgres_pool.Pool
}

func SetupPostgres(ctx context.Context, migrationDir string) (*PostgresContainer, error) {
	const (
		dbName = "testdb"
		dbUser = "testuser"
		dbPass = "testpass"
	)

	initScript, err := filepath.Glob(filepath.Join(migrationDir, "*up.sql"))
	if err != nil {
		return nil, fmt.Errorf("glob migration files: %w", err)
	}
	if len(initScript) == 0 {
		return nil, fmt.Errorf("no migration files found in %s", migrationDir)
	}
	sort.Strings(initScript)

	// testcontainers-go поднимает Postgres в Docker:
	// - WithDatabase/WithUsername/WithPassword задают креды
	// - WithInitScripts выполняет SQL-файлы при старте (аналог docker-entrypoint-initdb.d)
	// - wait.ForLog ждёт, пока Postgres не напишет "ready to accept connections"

	pgContainer, err := postgres.Run(
		ctx,
		"postgres:alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPass),
		postgres.WithInitScripts(initScript...),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("run postgres container: %w", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = pgContainer.Terminate(ctx)
		return nil, fmt.Errorf("get connection string: %w", err)
	}

	pool, err := core_pgx_pool.NewPoolFromConnString(ctx, connStr, 5*time.Second)
	if err != nil {
		// Если пул не создался — всё равно нужно остановить контейнер
		_ = pgContainer.Terminate(ctx)
		return nil, fmt.Errorf("create pool: %w", err)
	}

	return &PostgresContainer{
		Container: pgContainer,
		Pool:      pool,
	}, nil
}

// Teardown останавливает контейнер и закрывает пул.
func (pc *PostgresContainer) Teardown(ctx context.Context) {
	pc.Pool.Close()
	_ = pc.Container.Terminate(ctx)
}
