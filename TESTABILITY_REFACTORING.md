# Рефакторинг: подготовка проекта к написанию тестов

Описание изменений, выполненных перед добавлением тестов.
Цель — устранить архитектурные проблемы, из-за которых тесты пришлось бы переписывать позже.

---

## 1. Logger: добавлен `NewNopLogger()` и безопасный `FromContext()`

**Файл:** `internal/core/logger/logger.go`

### Что сделано

- Добавлена функция `NewNopLogger()` — создаёт логгер-заглушку (`zap.NewNop()`), который ничего не пишет и не требует файловой системы.
- Метод `Close()` теперь безопасен для NopLogger — проверяет `l.file == nil` перед закрытием файла.
- `FromContext()` больше **не паникует** при отсутствии логгера в контексте — возвращает `NewNopLogger()` как fallback.

### Зачем

- **`NewNopLogger()`**: без этого для тестирования любого хендлера, middleware или `HTTPResponseHandler` требовалось создавать реальный `*Logger` с файлом на диске. Теперь в тестах достаточно: `log := core_logger.NewNopLogger()`.
- **`FromContext()` без panic**: каждый HTTP-хендлер вызывает `core_logger.FromContext(ctx)`. Если в тесте забыть положить логгер в контекст — раньше тест падал с `panic("no log in context")` без полезной информации. Теперь вместо panic возвращается nop-логгер, и тест продолжает работать.

---

## 2. Domain: убран `time.Now()` из `NewTaskUninitialized()` и `ApplyPatch()`

**Файл:** `internal/core/domain/task.go`

### Что сделано

- `NewTaskUninitialized()` теперь принимает параметр `createdAt time.Time` вместо вызова `time.Now()` внутри.
- `ApplyPatch()` теперь принимает параметр `now time.Time` вместо вызова `time.Now()` внутри.
- Обновлены все вызывающие места:
  - `internal/features/tasks/transport/http/create_task.go` — передаёт `time.Now()`.
  - `internal/features/tasks/service/patch_task.go` — передаёт `time.Now()`.

### Зачем

Скрытый вызов `time.Now()` внутри доменной логики делал тесты **недетерминированными**:

- В тесте `NewTaskUninitialized(...)` нельзя было точно проверить `CreatedAt`.
- В тесте `ApplyPatch(completed=true)` нельзя было точно проверить `CompletedAt` — приходилось бы использовать приблизительные сравнения (`time.Since(got) < 1s`), которые ненадёжны.

Теперь в тестах можно передать фиксированное время:
```go
fixedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
task := domain.NewTaskUninitialized("title", nil, 1, fixedTime)
assert.Equal(t, fixedTime, task.CreatedAt)
```

---

## 3. WebService: убран `os.Getenv()` из сервис-слоя

**Файлы:**
- `internal/features/web/service/service.go`
- `internal/features/web/service/get_main_page.go`
- `cmd/todoapp/main.go`

### Что сделано

- В `WebService` добавлено поле `projectRoot string`.
- Конструктор `NewWebService()` теперь принимает `projectRoot` как параметр.
- `GetMainPage()` использует `s.projectRoot` вместо `os.Getenv("PROJECT_ROOT")`.
- В `main.go` значение `os.Getenv("PROJECT_ROOT")` передаётся в конструктор.

### Зачем

Прямой вызов `os.Getenv("PROJECT_ROOT")` в сервис-слое — это **скрытая зависимость**, нарушающая принцип инверсии зависимостей:

- В тесте пришлось бы вызывать `os.Setenv("PROJECT_ROOT", ...)` — это глобальное состояние, которое ломает параллельные тесты (`t.Parallel()`).
- Если позже конфигурация переедет в структуру `AppConfig`, тесты с `os.Setenv` сломаются.

Теперь в тестах зависимость инжектируется явно:
```go
svc := web_service.NewWebService(mockRepo, "/tmp/test-project")
```

---

## 4. Убрано дублирование валидации в transport-слое

**Файлы:**
- `internal/features/tasks/transport/http/patch_task.go`
- `internal/features/users/transport/http/patch_user.go`

### Что сделано

- Удалён метод `PatchTaskRequest.Validate()` — дублировал проверки из `domain.TaskPatch.Validate()` и `domain.Task.Validate()`.
- Удалён метод `PatchUserRequest.Validate()` — дублировал проверки из `domain.UserPatch.Validate()` и `domain.User.Validate()`.

### Зачем

Валидация была продублирована в двух слоях с **расхождением правил** (например, `PatchTaskRequest` проверяла `title` на длину 1–100, а `domain.Task.Validate()` — тоже 1–100, но `description` в транспорте вообще не проверялся). При написании тестов:

1. Пришлось бы писать тесты для обоих слоёв на одну и ту же логику.
2. При изменении бизнес-правил нужно было бы менять в двух местах и обновлять два набора тестов.
3. Правила уже расходились — значит тесты закодировали бы противоречивое поведение.

Теперь бизнес-валидация — **только в domain**. Transport-слой отвечает лишь за декодирование HTTP-запроса. Единый источник правды — один набор тестов.

---

## 5. Вынесен общий `TaskModel` из дублирующихся пакетов

**Файлы:**
- **Создан:** `internal/core/repository/postgres/model/task.go` — общий пакет.
- **Обновлены:**
  - `internal/features/tasks/repository/postgres/models.go` — использует type alias и ссылки на общий пакет.
  - `internal/features/statistics/repository/postgres/model.go` — аналогично.

### Что сделано

- `TaskModel`, `TaskDomainFromModel()`, `TaskDomainsFromModel()` вынесены в `core_postgres_model`.
- В пакетах `tasks` и `statistics` репозиториев оставлены type alias и var-ссылки для обратной совместимости — остальные файлы не потребовали изменений.

### Зачем

`TaskModel` и функции-конвертеры были **идентичны** в двух пакетах. При написании тестов:

1. Пришлось бы тестировать один и тот же код дважды.
2. При изменении структуры задачи (новое поле, изменение маппинга) — менять и тестировать в двух местах.

Теперь конвертеры определены один раз, и тесты пишутся только для `core_postgres_model`.

---

## Итог

| # | Изменение | Проблема для тестов | Файлы |
|---|-----------|-------------------|-------|
| 1 | `NewNopLogger()` + безопасный `FromContext()` | Невозможно unit-тестировать хендлеры без реального логгера и файловой системы | `logger.go` |
| 2 | `time.Now()` вынесен из domain | Недетерминированные тесты, нельзя точно проверить время | `task.go`, `create_task.go`, `patch_task.go` |
| 3 | `os.Getenv` → DI через конструктор | Скрытая зависимость от env, ломает `t.Parallel()` | `service.go`, `get_main_page.go`, `main.go` |
| 4 | Удалена дублированная валидация | Двойные тесты, расхождение правил между слоями | `patch_task.go`, `patch_user.go` |
| 5 | Общий `TaskModel` | Дублирование тестов, двойное сопровождение | `model/task.go`, `models.go`, `model.go` |

Проект проверен: `go build ./...` и `go vet ./...` проходят без ошибок.
