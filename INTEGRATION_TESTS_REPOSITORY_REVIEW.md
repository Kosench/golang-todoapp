# Review: `INTEGRATION_TESTS_REPOSITORY.md`

## Краткий вывод

Идея из `INTEGRATION_TESTS_REPOSITORY.md` в целом правильная: репозитории действительно стоит проверять integration-тестами на настоящем Postgres через `testcontainers-go`.

Но текущий файл больше похож на учебный черновик или план реализации, чем на финальное хорошее решение. Есть несколько архитектурных противоречий и мест, которые лучше переделать.

Главная проблема: в кодовой базе сами integration-тесты почти не реализованы. Файл:

```text
internal/features/users/repository/postgres/repository_integration_test.go
```

сейчас содержит только package/import/глобальную переменную `pool`, но не содержит ни `TestMain`, ни тестов.

Команда:

```bash
go test ./...
```

проходит успешно, но для repository integration-тестов выводит фактически:

```text
ok github.com/Kosench/golang-todoapp/internal/features/users/repository/postgres [no tests to run]
```

То есть описанная в `.md` логика пока не стала полноценными тестами.

## Что сделано хорошо

### 1. Правильно выбрана цель integration-тестов

Integration-тесты репозиториев должны проверять не бизнес-логику, а реальную работу SQL с настоящей БД:

- `INSERT ... RETURNING`
- `SELECT ... WHERE id=$1`
- `UPDATE ... WHERE version=$N`
- `DELETE` + `RowsAffected`
- foreign keys
- check constraints
- маппинг SQL-ошибок в domain/core errors

Это правильный уровень тестирования для repository-слоя.

### 2. `testcontainers-go` подходит для задачи

Использование `testcontainers-go` + Postgres-контейнера — хорошее решение:

- тесты не зависят от локальной БД разработчика;
- окружение воспроизводимое;
- можно запускать в CI;
- миграции применяются на чистой базе.

### 3. Build tag `integration` — правильная идея

В документе предлагается:

```go
//go:build integration
```

Это хорошая практика.

Обычные unit-тесты запускаются так:

```bash
go test ./...
```

А integration-тесты отдельно:

```bash
go test -tags=integration ./...
```

Это важно, потому что integration-тесты требуют Docker и работают дольше.

### 4. Хорошая идея добавить `NewPoolFromConnString`

В проекте уже есть конструктор:

```go
func NewPoolFromConnString(ctx context.Context, connStr string, opTimeout time.Duration) (*Pool, error)
```

Он находится в:

```text
internal/core/repository/postgres/pool/pgx/pool.go
```

Это хорошее решение для integration-тестов, потому что `testcontainers-go` умеет отдавать готовую connection string.

## Основные проблемы

### 1. Документ противоречит сам себе по структуре тестов

В документе написано, что рекомендуется общий integration-тестовый пакет:

```text
internal/integration_tests/
```

Но дальше примеры показывают `TestMain` внутри конкретных пакетов:

```text
internal/features/users/repository/postgres/repository_integration_test.go
internal/features/tasks/repository/postgres/repository_integration_test.go
internal/features/statistics/repository/postgres/repository_integration_test.go
```

В Go `TestMain` работает только в рамках одного package. Поэтому если тесты лежат в разных package, общий `TestMain` на все repository-тесты напрямую не получится.

Нужно выбрать один подход.

## Варианты структуры

### Вариант A: тесты рядом с репозиториями

Пример:

```text
internal/features/users/repository/postgres/repository_integration_test.go
internal/features/tasks/repository/postgres/repository_integration_test.go
internal/features/statistics/repository/postgres/repository_integration_test.go
```

Плюсы:

- тесты рядом с кодом;
- удобно запускать тесты конкретного repository package;
- привычная Go-структура.

Минусы:

- у каждого package будет свой `TestMain`;
- при `go test -tags=integration ./...` может подниматься несколько Postgres-контейнеров;
- будет больше дублирования setup-кода.

### Вариант B: отдельный общий integration package

Пример:

```text
internal/integration_test/
  testhelper/
    postgres.go
  repository/
    main_test.go
    users_repository_test.go
    tasks_repository_test.go
    statistics_repository_test.go
```

Все repository integration-тесты находятся в одном package, например:

```go
package repository_integration_test
```

Плюсы:

- один `TestMain`;
- один Postgres-контейнер;
- удобно тестировать связку `users/tasks/statistics`;
- меньше дублирования setup-кода.

Минусы:

- тесты лежат дальше от конкретной реализации repository;
- нужно аккуратно организовать helper-функции.

Для этого проекта лучше выглядит **вариант B**, потому что `tasks` и `statistics` зависят от `users`, и общий контейнер для всех repository-тестов логичен.

## Что улучшить в test helper

Сейчас есть helper:

```text
internal/integration_test/testhelper/postgres.go
```

Он работает в правильном направлении, но его можно упростить.

### 1. Использовать `NewPoolFromConnString`

Сейчас helper получает host и mapped port:

```go
host, err := pgContainer.Host(ctx)
port, err := pgContainer.MappedPort(ctx, "5432/tcp")
```

После этого вручную создаётся config:

```go
pool, err := core_pgx_pool.NewPool(ctx, core_pgx_pool.Config{
    Host:     host,
    Port:     port.Port(),
    User:     dbUser,
    Pass:     dbPass,
    Database: dbName,
    Timeout:  5 * time.Second,
})
```

Проще и лучше:

```go
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
```

Так меньше ручной сборки DSN и меньше мест для ошибки.

### 2. Проверять, что миграции найдены

Сейчас используется:

```go
initScript, err := filepath.Glob(filepath.Join(migrationDir, "*up.sql"))
if err != nil {
    return nil, fmt.Errorf("glob migration files: %w", err)
}
```

Проблема: если путь к миграциям неправильный, `Glob` вернёт пустой список без ошибки. Контейнер поднимется без таблиц, а тесты упадут позже с неочевидной SQL-ошибкой.

Лучше:

```go
initScripts, err := filepath.Glob(filepath.Join(migrationDir, "*up.sql"))
if err != nil {
    return nil, fmt.Errorf("glob migration files: %w", err)
}
if len(initScripts) == 0 {
    return nil, fmt.Errorf("no migration files found in %s", migrationDir)
}
```

### 3. Явно сортировать миграции

`filepath.Glob` обычно возвращает отсортированный список, но для явности лучше добавить:

```go
sort.Strings(initScripts)
```

Для миграций порядок критичен.

## Что улучшить в самих тестах

### 1. Не делать тесты зависимыми от порядка subtests

В документе показан CRUD-тест, где subtests идут последовательно:

```go
t.Run("CreateUser", ...)
t.Run("GetUser", ...)
t.Run("PatchUser", ...)
t.Run("DeleteUser", ...)
t.Run("DeleteUser_NotFound", ...)
```

И следующие subtests используют данные из предыдущих.

Это работает, но неидеально:

- если `CreateUser` упал, остальные тесты падают каскадом;
- отдельный сценарий сложнее отлаживать;
- тесты нельзя безопасно распараллелить;
- состояние теста размазано по нескольким subtests.

Лучше делать отдельные независимые тесты:

```go
func TestUsersRepository_CreateUser(t *testing.T) {}
func TestUsersRepository_GetUser(t *testing.T) {}
func TestUsersRepository_PatchUser(t *testing.T) {}
func TestUsersRepository_DeleteUser(t *testing.T) {}
```

Каждый тест сам создаёт нужные данные.

### 2. Чистить БД между тестами

Если все тесты используют один контейнер и одну БД, лучше очищать таблицы перед каждым тестом.

Например:

```go
func truncateTables(ctx context.Context, pool core_postgres_pool.Pool) error {
    _, err := pool.Exec(ctx, `
        TRUNCATE TABLE tasks, users RESTART IDENTITY CASCADE
    `)
    return err
}
```

И helper:

```go
func cleanup(t *testing.T) {
    t.Helper()

    if err := truncateTables(context.Background(), pool); err != nil {
        t.Fatalf("truncate tables: %v", err)
    }
}
```

Тогда каждый тест начинается с чистого состояния.

### 3. Усилить проверки `GetUsers` / `GetTasks`

В документе есть проверка такого типа:

```go
if len(users) < 1 {
    t.Error("expected at least 1 user")
}
```

Она слишком слабая. Такой тест может пройти, даже если вернулся не тот пользователь.

Лучше:

```go
found := false
for _, user := range users {
    if user.ID == createdUser.ID {
        found = true
        break
    }
}
if !found {
    t.Fatalf("created user %d not found in GetUsers result", createdUser.ID)
}
```

А если БД очищается перед тестом, можно проверять точное количество:

```go
if len(users) != 1 {
    t.Fatalf("len(users) = %d, want 1", len(users))
}
```

### 4. Не использовать магический `999999`

В примерах часто используется:

```go
repo.GetUser(ctx, 999999)
```

Лучше сделать константу:

```go
const missingID = 999999999
```

Так понятнее, что это намеренно несуществующий ID.

### 5. Аккуратнее обрабатывать ошибки в `TestMain`

В документе используется `panic`:

```go
panic("failed to start postgres container: " + err.Error())
```

Для `TestMain` лучше использовать:

```go
fmt.Fprintf(os.Stderr, "setup postgres: %v\n", err)
os.Exit(1)
```

Вывод будет чище и понятнее.

## Рекомендуемая структура для проекта

Оптимальный вариант:

```text
internal/integration_test/
  testhelper/
    postgres.go
  repository/
    main_test.go
    users_repository_test.go
    tasks_repository_test.go
    statistics_repository_test.go
```

### `main_test.go`

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

### Пример независимого теста

```go
func TestUsersRepository_GetUser(t *testing.T) {
    cleanup(t)

    ctx := context.Background()
    repo := users_postgres_repository.NewUsersRepository(pool)

    created, err := repo.CreateUser(ctx, domain.NewUserUninitialized("John Doe", ptr("+1234567890")))
    if err != nil {
        t.Fatalf("create user: %v", err)
    }

    got, err := repo.GetUser(ctx, created.ID)
    if err != nil {
        t.Fatalf("get user: %v", err)
    }

    if got.ID != created.ID {
        t.Errorf("ID = %d, want %d", got.ID, created.ID)
    }
    if got.FullName != "John Doe" {
        t.Errorf("FullName = %q, want %q", got.FullName, "John Doe")
    }
}
```

## Итоговая оценка

### Идея

**8/10**

Выбран правильный тип тестирования и правильный инструмент.

### Документ

**6/10**

Документ полезен как объяснение и черновой план, но содержит противоречия по структуре тестов и не всегда показывает лучший test design.

### Текущая реализация в коде

**2/10**

Пока настоящих integration-тестов практически нет. Есть helper и добавленный конструктор `NewPoolFromConnString`, но сами repository-тесты из markdown не перенесены в `_test.go`.

## Что сделать первым

1. Выбрать структуру: общий integration package или тесты рядом с repository.
2. Лучше выбрать общий package:

   ```text
   internal/integration_test/repository/
   ```

3. Переделать `testhelper.SetupPostgres` на `NewPoolFromConnString`.
4. Добавить проверку, что миграции найдены.
5. Добавить очистку БД между тестами.
6. Сделать каждый тест независимым.
7. Перенести реальные тесты из markdown в `_test.go`.
