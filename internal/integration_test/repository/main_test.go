//go:build integration

package repository_integration_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	core_postgres_pool "github.com/Kosench/golang-todoapp/internal/core/repository/postgres/pool"
	"github.com/Kosench/golang-todoapp/internal/integration_test/testhelper"
)

var pool core_postgres_pool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	migrationsDir, err := filepath.Abs("../../../migrations")
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve migrations dir: %v\n", err)
		os.Exit(1)
	}

	pg, err := testhelper.SetupPostgres(ctx, migrationsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "setup postgres: %v\n", err)
		os.Exit(1)
	}

	pool = pg.Pool
	code := m.Run()
	pg.Teardown(ctx)

	os.Exit(code)
}
