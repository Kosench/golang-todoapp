# golang-todoapp

REST API для todo-приложения на Go. Пет-проект для практики Go, чистой архитектуры и работы с PostgreSQL.

## Стек

- Go 1.25
- PostgreSQL
- net/http + кастомный маршрутизатор
- Swagger (документация API)
- Docker Compose (окружение)
- zap (логирование)
- envconfig (конфигурация)
- validator (валидация)

## Архитектура

Проект организован по features с чистой архитектурой:

```
internal/
├── core/                    # общие компоненты
│   ├── domain/              # доменные модели
│   ├── transport/http/      # HTTP сервер, middleware
│   ├── repository/          # пул соединений
│   └── logger/
└── features/
    ├── users/
    ├── tasks/
    └── statistics/
        ├── transport/       # HTTP handlers
        ├── service/         # бизнес-логика
        └── repository/      # слой данных
```

Каждая feature содержит: transport → service → repository.

## API

Документация доступна через Swagger UI по адресу `/swagger/` после запуска приложения.

### Endpoints

**Users**
- `POST /api/v1/users` — создать пользователя
- `GET /api/v1/users` — список пользователей
- `GET /api/v1/users/{id}` — получить пользователя
- `PATCH /api/v1/users/{id}` — обновить пользователя
- `DELETE /api/v1/users/{id}` — удалить пользователя

**Tasks**
- `POST /api/v1/tasks` — создать задачу
- `GET /api/v1/tasks` — список задач (фильтр по `user_id`, пагинация `limit`/`offset`)
- `GET /api/v1/tasks/{id}` — получить задачу
- `PATCH /api/v1/tasks/{id}` — частично обновить задачу
- `DELETE /api/v1/tasks/{id}` — удалить задачу

**Statistics**
- `GET /api/v1/statistics` — статистика по задачам (фильтр по `user_id`, `from`, `to`)

## Запуск

### 1. Подготовка окружения

Скопируй `.env.example` в `.env` и заполни:

```bash
cp .env.example .env
```

```
POSTGRES_USER=todoapp
POSTGRES_PASSWORD=secret
POSTGRES_DB=todoapp
```

### 2. Запуск PostgreSQL

```bash
make env-up
```

### 3. Миграции

```bash
make migrate-up
```

### 4. Запуск приложения

Локально:

```bash
make todoapp-run
```

Или через Docker:

```bash
make todoapp-deploy
```

Swagger UI: `http://127.0.0.1:5050/swagger/`

### Остановка

```bash
make env-down          # остановить БД
make env-cleanup       # удалить volumes
make todoapp-undeploy  # остановить приложение
```

## Swagger

Генерация документации:

```bash
make swagger-gen
```

## Перегенерация Swagger

При изменении handler-ов нужно перегенерировать swagger-файлы:

```bash
make swagger-gen
```

## Логирование

По умолчанию логи пишутся в `out/log/`. Для очистки:

```bash
make logs-cleanup
```

## Port forward к БД

Если нужно подключиться к БД из хоста:

```bash
make env-port-forward   # открыть проброс на 127.0.0.1:5432
make env-port-close     # закрыть
```

## Что можно улучшить

- [ ] Авторизация/аутентификация
- [ ] Тесты
- [ ] GraphQL endpoint
- [ ] Кеширование (Redis)
- [ ] CI/CD
