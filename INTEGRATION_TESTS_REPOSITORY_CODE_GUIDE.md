# Code guide: integration-тесты repository-слоя

Этот файл — не код для автоматического применения. Это набор примеров, которые можно вручную перенести в проект.

Цель:

1. Выбрать общий integration package.
2. Переделать `testhelper.SetupPostgres` на `NewPoolFromConnString`.
3. Добавить проверку наличия миграций.
4. Добавить очистку БД между тестами.
5. Сделать тесты независимыми.
6. Перенести реальные тесты из `INTEGRATION_TESTS_REPOSITORY.md` в `_test.go`.

## 1. Рекомендуемая структура

Создай такую структуру:

```text
internal/integration_test/
  testhelper/
    postgres.go
  repository/
    main_test.go
    helpers_test.go
    users_repository_test.go
    tasks_repository_test.go
    statistics_repository_test.go
```

Все файлы в `internal/integration_test/repository/` должны быть в одном package:

```go
package repository_integration_test
```

Так у всех repository integration-тестов будет один общий `TestMain` и один Postgres-контейнер.

## 2. `testhelper/postgres.go`

Файл:

```text
internal/integration_test/testhelper/postgres.go
```

Пример кода:

```go
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

type PostgresContainer struct {
	Container testcontainers.Container
	Pool      core_postgres_pool.Pool
}

func SetupPostgres(ctx context.Context, migrationsDir string) (*PostgresContainer, error) {
	const (
		dbName = "testdb"
		dbUser = "testuser"
		dbPass = "testpass"
	)

	initScripts, err := filepath.Glob(filepath.Join(migrationsDir, "*up.sql"))
	if err != nil {
		return nil, fmt.Errorf("glob migration files: %w", err)
	}
	if len(initScripts) == 0 {
		return nil, fmt.Errorf("no migration files found in %s", migrationsDir)
	}
	sort.Strings(initScripts)

	pgContainer, err := postgres.Run(
		ctx,
		"postgres:alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPass),
		postgres.WithInitScripts(initScripts...),
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
		_ = pgContainer.Terminate(ctx)
		return nil, fmt.Errorf("create pool: %w", err)
	}

	return &PostgresContainer{
		Container: pgContainer,
		Pool:      pool,
	}, nil
}

func (pc *PostgresContainer) Teardown(ctx context.Context) {
	pc.Pool.Close()
	_ = pc.Container.Terminate(ctx)
}
```

Короткое пояснение:

- `filepath.Glob(... "*up.sql")` ищет миграции.
- `len(initScripts) == 0` нужен, чтобы быстро поймать неправильный путь к миграциям.
- `sort.Strings(initScripts)` делает порядок миграций явным.
- `ConnectionString` берёт готовый DSN у testcontainers.
- `NewPoolFromConnString` создаёт твой проектный pool без ручной сборки host/port/user/pass.

## 3. `repository/main_test.go`

Файл:

```text
internal/integration_test/repository/main_test.go
```

Пример кода:

```go
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
```

Короткое пояснение:

- `//go:build integration` исключает файл из обычного `go test ./...`.
- `TestMain` поднимает контейнер один раз на весь package.
- Путь `../../../migrations` считается от директории `internal/integration_test/repository/`.

Запуск:

```bash
go test -tags=integration ./internal/integration_test/repository
```

Или все тесты проекта вместе с integration:

```bash
go test -tags=integration ./...
```

## 4. `repository/helpers_test.go`

Файл:

```text
internal/integration_test/repository/helpers_test.go
```

Пример кода:

```go
//go:build integration

package repository_integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
	users_postgres_repository "github.com/Kosench/golang-todoapp/internal/features/users/repository/postgres"
)

const missingID = 999999999

func ptr[T any](v T) *T {
	return &v
}

func cleanup(t *testing.T) {
	t.Helper()

	_, err := pool.Exec(context.Background(), `
		TRUNCATE TABLE todoapp.tasks, todoapp.users RESTART IDENTITY CASCADE
	`)
	if err != nil {
		t.Fatalf("cleanup database: %v", err)
	}
}

func createUser(t *testing.T, ctx context.Context, fullName string) domain.User {
	t.Helper()

	repo := users_postgres_repository.NewUsersRepository(pool)

	user, err := repo.CreateUser(ctx, domain.NewUserUninitialized(fullName, ptr("+1234567890")))
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	return user
}

func fixedTime() time.Time {
	return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
}
```

Короткое пояснение:

- `cleanup` чистит таблицы перед каждым тестом.
- `RESTART IDENTITY` сбрасывает auto-increment.
- `CASCADE` нужен из-за foreign key `tasks.author_user_id -> users.id`.
- `createUser` нужен, потому что task tests зависят от существующего пользователя.
- `fixedTime` лучше, чем `time.Now()`, потому что тесты становятся стабильнее.

## 5. `users_repository_test.go`

Файл:

```text
internal/integration_test/repository/users_repository_test.go
```

Пример кода:

```go
//go:build integration

package repository_integration_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
	core_errors "github.com/Kosench/golang-todoapp/internal/core/errors"
	users_postgres_repository "github.com/Kosench/golang-todoapp/internal/features/users/repository/postgres"
)

func TestUsersRepository_CreateUser(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	repo := users_postgres_repository.NewUsersRepository(pool)

	user, err := repo.CreateUser(ctx, domain.NewUserUninitialized("John Doe", ptr("+1234567890")))
	if err != nil {
		t.Fatalf("CreateUser() error: %v", err)
	}

	if user.ID <= 0 {
		t.Errorf("ID = %d, want positive", user.ID)
	}
	if user.Version != 1 {
		t.Errorf("Version = %d, want 1", user.Version)
	}
	if user.FullName != "John Doe" {
		t.Errorf("FullName = %q, want %q", user.FullName, "John Doe")
	}
	if user.PhoneNumber == nil || *user.PhoneNumber != "+1234567890" {
		t.Errorf("PhoneNumber = %v, want %q", user.PhoneNumber, "+1234567890")
	}
}

func TestUsersRepository_GetUser(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	repo := users_postgres_repository.NewUsersRepository(pool)
	created := createUser(t, ctx, "John Doe")

	got, err := repo.GetUser(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetUser() error: %v", err)
	}

	if got.ID != created.ID {
		t.Errorf("ID = %d, want %d", got.ID, created.ID)
	}
	if got.FullName != created.FullName {
		t.Errorf("FullName = %q, want %q", got.FullName, created.FullName)
	}
}

func TestUsersRepository_GetUser_NotFound(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	repo := users_postgres_repository.NewUsersRepository(pool)

	_, err := repo.GetUser(ctx, missingID)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestUsersRepository_GetUsers(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	repo := users_postgres_repository.NewUsersRepository(pool)
	first := createUser(t, ctx, "John Doe")
	second := createUser(t, ctx, "Jane Doe")

	users, err := repo.GetUsers(ctx, ptr(10), ptr(0))
	if err != nil {
		t.Fatalf("GetUsers() error: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("len(users) = %d, want 2", len(users))
	}
	if users[0].ID != first.ID {
		t.Errorf("users[0].ID = %d, want %d", users[0].ID, first.ID)
	}
	if users[1].ID != second.ID {
		t.Errorf("users[1].ID = %d, want %d", users[1].ID, second.ID)
	}
}

func TestUsersRepository_PatchUser(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	repo := users_postgres_repository.NewUsersRepository(pool)
	created := createUser(t, ctx, "John Doe")

	patched := created
	patched.FullName = "Jane Doe"
	patched.PhoneNumber = nil

	got, err := repo.PatchUser(ctx, created.ID, patched)
	if err != nil {
		t.Fatalf("PatchUser() error: %v", err)
	}

	if got.FullName != "Jane Doe" {
		t.Errorf("FullName = %q, want %q", got.FullName, "Jane Doe")
	}
	if got.PhoneNumber != nil {
		t.Errorf("PhoneNumber = %v, want nil", got.PhoneNumber)
	}
	if got.Version != created.Version+1 {
		t.Errorf("Version = %d, want %d", got.Version, created.Version+1)
	}
}

func TestUsersRepository_PatchUser_Conflict(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	repo := users_postgres_repository.NewUsersRepository(pool)
	created := createUser(t, ctx, "John Doe")

	firstPatch := created
	firstPatch.FullName = "Jane Doe"

	updated, err := repo.PatchUser(ctx, created.ID, firstPatch)
	if err != nil {
		t.Fatalf("first PatchUser() error: %v", err)
	}

	stale := created
	stale.FullName = "Stale Name"

	_, err = repo.PatchUser(ctx, updated.ID, stale)
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("error = %v, want ErrConflict", err)
	}
}

func TestUsersRepository_DeleteUser(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	repo := users_postgres_repository.NewUsersRepository(pool)
	created := createUser(t, ctx, "John Doe")

	if err := repo.DeleteUser(ctx, created.ID); err != nil {
		t.Fatalf("DeleteUser() error: %v", err)
	}

	_, err := repo.GetUser(ctx, created.ID)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("GetUser() after delete error = %v, want ErrNotFound", err)
	}
}

func TestUsersRepository_DeleteUser_NotFound(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	repo := users_postgres_repository.NewUsersRepository(pool)

	err := repo.DeleteUser(ctx, missingID)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}
```

Короткое пояснение:

- Каждый тест вызывает `cleanup(t)`.
- Каждый тест сам создаёт нужные данные.
- Нет зависимости между `Create`, `Get`, `Patch`, `Delete`.
- `PatchUser_Conflict` сначала реально повышает version, потом пытается сохранить старую version.

## 6. `tasks_repository_test.go`

Файл:

```text
internal/integration_test/repository/tasks_repository_test.go
```

Пример кода:

```go
//go:build integration

package repository_integration_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
	core_errors "github.com/Kosench/golang-todoapp/internal/core/errors"
	tasks_postgres_repository "github.com/Kosench/golang-todoapp/internal/features/tasks/repository/postgres"
)

func TestTasksRepository_CreateTask(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	author := createUser(t, ctx, "Task Author")
	repo := tasks_postgres_repository.NewTasksRepository(pool)

	task := domain.NewTaskUninitialized("Buy milk", ptr("Two bottles"), author.ID, fixedTime())

	created, err := repo.CreateTask(ctx, task)
	if err != nil {
		t.Fatalf("CreateTask() error: %v", err)
	}

	if created.ID <= 0 {
		t.Errorf("ID = %d, want positive", created.ID)
	}
	if created.Version != 1 {
		t.Errorf("Version = %d, want 1", created.Version)
	}
	if created.Title != "Buy milk" {
		t.Errorf("Title = %q, want %q", created.Title, "Buy milk")
	}
	if created.Description == nil || *created.Description != "Two bottles" {
		t.Errorf("Description = %v, want %q", created.Description, "Two bottles")
	}
	if created.Completed {
		t.Errorf("Completed = true, want false")
	}
	if created.AuthorUserID != author.ID {
		t.Errorf("AuthorUserID = %d, want %d", created.AuthorUserID, author.ID)
	}
}

func TestTasksRepository_CreateTask_ForeignKeyViolation(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	repo := tasks_postgres_repository.NewTasksRepository(pool)

	task := domain.NewTaskUninitialized("Bad task", nil, missingID, fixedTime())

	_, err := repo.CreateTask(ctx, task)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestTasksRepository_GetTask(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	author := createUser(t, ctx, "Task Author")
	repo := tasks_postgres_repository.NewTasksRepository(pool)

	created, err := repo.CreateTask(ctx, domain.NewTaskUninitialized("Buy milk", nil, author.ID, fixedTime()))
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	got, err := repo.GetTask(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetTask() error: %v", err)
	}

	if got.ID != created.ID {
		t.Errorf("ID = %d, want %d", got.ID, created.ID)
	}
	if got.Title != created.Title {
		t.Errorf("Title = %q, want %q", got.Title, created.Title)
	}
}

func TestTasksRepository_GetTask_NotFound(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	repo := tasks_postgres_repository.NewTasksRepository(pool)

	_, err := repo.GetTask(ctx, missingID)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestTasksRepository_GetTasks(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	author := createUser(t, ctx, "Task Author")
	repo := tasks_postgres_repository.NewTasksRepository(pool)

	first, err := repo.CreateTask(ctx, domain.NewTaskUninitialized("First task", nil, author.ID, fixedTime()))
	if err != nil {
		t.Fatalf("create first task: %v", err)
	}
	second, err := repo.CreateTask(ctx, domain.NewTaskUninitialized("Second task", nil, author.ID, fixedTime()))
	if err != nil {
		t.Fatalf("create second task: %v", err)
	}

	tasks, err := repo.GetTasks(ctx, nil, ptr(10), ptr(0))
	if err != nil {
		t.Fatalf("GetTasks() error: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("len(tasks) = %d, want 2", len(tasks))
	}
	if tasks[0].ID != first.ID {
		t.Errorf("tasks[0].ID = %d, want %d", tasks[0].ID, first.ID)
	}
	if tasks[1].ID != second.ID {
		t.Errorf("tasks[1].ID = %d, want %d", tasks[1].ID, second.ID)
	}
}

func TestTasksRepository_GetTasks_FilterByUser(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	firstAuthor := createUser(t, ctx, "First Author")
	secondAuthor := createUser(t, ctx, "Second Author")
	repo := tasks_postgres_repository.NewTasksRepository(pool)

	firstTask, err := repo.CreateTask(ctx, domain.NewTaskUninitialized("First author task", nil, firstAuthor.ID, fixedTime()))
	if err != nil {
		t.Fatalf("create first author task: %v", err)
	}
	_, err = repo.CreateTask(ctx, domain.NewTaskUninitialized("Second author task", nil, secondAuthor.ID, fixedTime()))
	if err != nil {
		t.Fatalf("create second author task: %v", err)
	}

	tasks, err := repo.GetTasks(ctx, &firstAuthor.ID, ptr(10), ptr(0))
	if err != nil {
		t.Fatalf("GetTasks() error: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(tasks))
	}
	if tasks[0].ID != firstTask.ID {
		t.Errorf("tasks[0].ID = %d, want %d", tasks[0].ID, firstTask.ID)
	}
	if tasks[0].AuthorUserID != firstAuthor.ID {
		t.Errorf("AuthorUserID = %d, want %d", tasks[0].AuthorUserID, firstAuthor.ID)
	}
}

func TestTasksRepository_PatchTask(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	author := createUser(t, ctx, "Task Author")
	repo := tasks_postgres_repository.NewTasksRepository(pool)

	created, err := repo.CreateTask(ctx, domain.NewTaskUninitialized("Old title", nil, author.ID, fixedTime()))
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	completedAt := fixedTime().Add(time.Hour)
	patched := created
	patched.Title = "New title"
	patched.Completed = true
	patched.CompletedAt = &completedAt

	got, err := repo.PatchTask(ctx, created.ID, patched)
	if err != nil {
		t.Fatalf("PatchTask() error: %v", err)
	}

	if got.Title != "New title" {
		t.Errorf("Title = %q, want %q", got.Title, "New title")
	}
	if !got.Completed {
		t.Errorf("Completed = false, want true")
	}
	if got.CompletedAt == nil {
		t.Fatalf("CompletedAt = nil, want non-nil")
	}
	if got.Version != created.Version+1 {
		t.Errorf("Version = %d, want %d", got.Version, created.Version+1)
	}
}

func TestTasksRepository_PatchTask_Conflict(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	author := createUser(t, ctx, "Task Author")
	repo := tasks_postgres_repository.NewTasksRepository(pool)

	created, err := repo.CreateTask(ctx, domain.NewTaskUninitialized("Old title", nil, author.ID, fixedTime()))
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	firstPatch := created
	firstPatch.Title = "New title"

	updated, err := repo.PatchTask(ctx, created.ID, firstPatch)
	if err != nil {
		t.Fatalf("first PatchTask() error: %v", err)
	}

	stale := created
	stale.Title = "Stale title"

	_, err = repo.PatchTask(ctx, updated.ID, stale)
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("error = %v, want ErrConflict", err)
	}
}

func TestTasksRepository_DeleteTask(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	author := createUser(t, ctx, "Task Author")
	repo := tasks_postgres_repository.NewTasksRepository(pool)

	created, err := repo.CreateTask(ctx, domain.NewTaskUninitialized("Task to delete", nil, author.ID, fixedTime()))
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	if err := repo.DeleteTask(ctx, created.ID); err != nil {
		t.Fatalf("DeleteTask() error: %v", err)
	}

	_, err = repo.GetTask(ctx, created.ID)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("GetTask() after delete error = %v, want ErrNotFound", err)
	}
}

func TestTasksRepository_DeleteTask_NotFound(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	repo := tasks_postgres_repository.NewTasksRepository(pool)

	err := repo.DeleteTask(ctx, missingID)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}
```

Важно: в этом файле используется `time.Hour`, поэтому в imports должен быть `time`.

Правильный import block для `tasks_repository_test.go`:

```go
import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
	core_errors "github.com/Kosench/golang-todoapp/internal/core/errors"
	tasks_postgres_repository "github.com/Kosench/golang-todoapp/internal/features/tasks/repository/postgres"
)
```

Короткое пояснение:

- Для task нужно сначала создать пользователя, потому что есть foreign key.
- `CreateTask_ForeignKeyViolation` проверяет, что несуществующий `author_user_id` мапится в `ErrNotFound`.
- `PatchTask_Conflict` проверяет optimistic locking через старую version.

## 7. `statistics_repository_test.go`

Файл:

```text
internal/integration_test/repository/statistics_repository_test.go
```

Пример кода:

```go
//go:build integration

package repository_integration_test

import (
	"context"
	"testing"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
	statistics_postgres_repository "github.com/Kosench/golang-todoapp/internal/features/statistics/repository/postgres"
	tasks_postgres_repository "github.com/Kosench/golang-todoapp/internal/features/tasks/repository/postgres"
)

func TestStatisticsRepository_GetTasks_All(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	author := createUser(t, ctx, "Stats Author")
	tasksRepo := tasks_postgres_repository.NewTasksRepository(pool)
	statsRepo := statistics_postgres_repository.NewStatisticsRepository(pool)

	first, err := tasksRepo.CreateTask(ctx, domain.NewTaskUninitialized("First task", nil, author.ID, fixedTime()))
	if err != nil {
		t.Fatalf("create first task: %v", err)
	}
	second, err := tasksRepo.CreateTask(ctx, domain.NewTaskUninitialized("Second task", nil, author.ID, fixedTime().Add(time.Hour)))
	if err != nil {
		t.Fatalf("create second task: %v", err)
	}

	tasks, err := statsRepo.GetTasks(ctx, nil, nil, nil)
	if err != nil {
		t.Fatalf("GetTasks() error: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("len(tasks) = %d, want 2", len(tasks))
	}
	if tasks[0].ID != first.ID {
		t.Errorf("tasks[0].ID = %d, want %d", tasks[0].ID, first.ID)
	}
	if tasks[1].ID != second.ID {
		t.Errorf("tasks[1].ID = %d, want %d", tasks[1].ID, second.ID)
	}
}

func TestStatisticsRepository_GetTasks_FilterByUser(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	firstAuthor := createUser(t, ctx, "First Author")
	secondAuthor := createUser(t, ctx, "Second Author")
	tasksRepo := tasks_postgres_repository.NewTasksRepository(pool)
	statsRepo := statistics_postgres_repository.NewStatisticsRepository(pool)

	firstTask, err := tasksRepo.CreateTask(ctx, domain.NewTaskUninitialized("First author task", nil, firstAuthor.ID, fixedTime()))
	if err != nil {
		t.Fatalf("create first author task: %v", err)
	}
	_, err = tasksRepo.CreateTask(ctx, domain.NewTaskUninitialized("Second author task", nil, secondAuthor.ID, fixedTime()))
	if err != nil {
		t.Fatalf("create second author task: %v", err)
	}

	tasks, err := statsRepo.GetTasks(ctx, &firstAuthor.ID, nil, nil)
	if err != nil {
		t.Fatalf("GetTasks() error: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(tasks))
	}
	if tasks[0].ID != firstTask.ID {
		t.Errorf("tasks[0].ID = %d, want %d", tasks[0].ID, firstTask.ID)
	}
}

func TestStatisticsRepository_GetTasks_FilterByDateRange(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	author := createUser(t, ctx, "Stats Author")
	tasksRepo := tasks_postgres_repository.NewTasksRepository(pool)
	statsRepo := statistics_postgres_repository.NewStatisticsRepository(pool)

	baseTime := fixedTime()
	beforeRange := baseTime.Add(-time.Hour)
	inRange := baseTime
	afterRange := baseTime.Add(time.Hour)

	_, err := tasksRepo.CreateTask(ctx, domain.NewTaskUninitialized("Before range", nil, author.ID, beforeRange))
	if err != nil {
		t.Fatalf("create before range task: %v", err)
	}
	expected, err := tasksRepo.CreateTask(ctx, domain.NewTaskUninitialized("In range", nil, author.ID, inRange))
	if err != nil {
		t.Fatalf("create in range task: %v", err)
	}
	_, err = tasksRepo.CreateTask(ctx, domain.NewTaskUninitialized("After range", nil, author.ID, afterRange))
	if err != nil {
		t.Fatalf("create after range task: %v", err)
	}

	from := baseTime.Add(-time.Minute)
	to := baseTime.Add(time.Minute)

	tasks, err := statsRepo.GetTasks(ctx, nil, &from, &to)
	if err != nil {
		t.Fatalf("GetTasks() error: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(tasks))
	}
	if tasks[0].ID != expected.ID {
		t.Errorf("tasks[0].ID = %d, want %d", tasks[0].ID, expected.ID)
	}
}
```

Важно: в этом файле используется `time.Hour` и `time.Minute`, поэтому в imports должен быть `time`.

Правильный import block:

```go
import (
	"context"
	"testing"
	"time"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
	statistics_postgres_repository "github.com/Kosench/golang-todoapp/internal/features/statistics/repository/postgres"
	tasks_postgres_repository "github.com/Kosench/golang-todoapp/internal/features/tasks/repository/postgres"
)
```

Короткое пояснение:

- Statistics repository читает tasks, поэтому данные создаются через `tasksRepo`.
- Проверяются основные фильтры: без фильтров, по userID, по диапазону дат.
- Диапазон дат в repository устроен так:

```sql
created_at >= from
created_at < to
```

Поэтому `to` не включается в результат.

## 8. Что удалить или не использовать

Если выбираешь общий package `internal/integration_test/repository/`, тогда файл:

```text
internal/features/users/repository/postgres/repository_integration_test.go
```

лучше удалить или оставить пустым не нужно.

Но удалять файл стоит только когда новые тесты уже перенесены и проходят.

Также не нужно делать отдельный `TestMain` в каждом repository package.

## 9. Как запускать

Обычные unit-тесты:

```bash
go test ./...
```

Integration-тесты repository:

```bash
go test -tags=integration ./internal/integration_test/repository
```

Все тесты с integration:

```bash
go test -tags=integration ./...
```

## 10. Возможные ошибки при переносе

### Ошибка: `undefined: time`

Значит в файле используется `time.Hour`, `time.Minute` или `time.Time`, но нет import:

```go
import "time"
```

### Ошибка: миграции не найдены

Проверь путь в `main_test.go`:

```go
migrationsDir, err := filepath.Abs("../../../migrations")
```

Этот путь правильный, если файл лежит тут:

```text
internal/integration_test/repository/main_test.go
```

### Ошибка Docker/testcontainers

Integration-тесты требуют запущенный Docker.

### Ошибка foreign key при очистке

Убедись, что cleanup чистит таблицы в правильной схеме:

```sql
TRUNCATE TABLE todoapp.tasks, todoapp.users RESTART IDENTITY CASCADE
```

## 11. Минимальный порядок ручного переноса

1. Обновить `internal/integration_test/testhelper/postgres.go`.
2. Создать директорию:

   ```text
   internal/integration_test/repository/
   ```

3. Создать `main_test.go`.
4. Создать `helpers_test.go`.
5. Перенести только `users_repository_test.go`.
6. Запустить:

   ```bash
   go test -tags=integration ./internal/integration_test/repository
   ```

7. Если users-тесты проходят, перенести `tasks_repository_test.go`.
8. Потом перенести `statistics_repository_test.go`.
9. Только после этого удалить старый пустой/лишний `repository_integration_test.go` рядом с users repository.
