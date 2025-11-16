# PR Reviewer Service

Сервис автоматического назначения ревьюверов на Pull Request с балансировкой нагрузки и отслеживанием статистики.

## Описание

Сервис решает задачу справедливого распределения code review между участниками команды:
- Автоматически назначает 2 ревьюверов на каждый PR
- Балансирует нагрузку (выбирает участников с наименьшим количеством открытых PR)
- Безопасно переназначает ревьюверов при деактивации пользователей
- Собирает статистику по командам и пользователям

## Быстрый старт

Клонирование и запуск
```bash
git clone https://github.com/T1mof/pr-reviewer-service.git
cd pr-reviewer-service
docker-compose up --build
```

Проверка работоспособности
```bash
curl http://localhost:8080/health

{"status":"ok"}
```

## API Endpoints

### Команды
- `POST /team/add` - Создание команды с участниками
- `GET /team/get?team_name={name}` - Получение информации о команде

### Пользователи
- `POST /users/setIsActive` - Деактивация/активация пользователя (требует X-Admin-Token)
- `GET /users/getReview?user_id={id}` - Список PR для ревью

### Pull Requests
- `POST /pullRequest/create` - Создание PR с автоназначением ревьюверов
- `POST /pullRequest/merge` - Закрытие PR
- `POST /pullRequest/reassign` - Переназначение ревьювера

### Статистика
- `GET /stats` - Общая статистика сервиса

## Примеры использования

### Создание команды
```bash
curl -X POST http://localhost:8080/team/add
-H "Content-Type: application/json"
-d '{
"team_name": "backend",
"members": [
{"user_id": "550e8400-e29b-41d4-a716-446655440001", "username": "Alice", "is_active": true},
{"user_id": "550e8400-e29b-41d4-a716-446655440002", "username": "Bob", "is_active": true},
{"user_id": "550e8400-e29b-41d4-a716-446655440003", "username": "Charlie", "is_active": true}
]
}'
```

### Создание Pull Request
```bash
curl -X POST http://localhost:8080/pullRequest/create
-H "Content-Type: application/json"
-d '{
"pull_request_id": "650e8400-e29b-41d4-a716-446655440001",
"pull_request_name": "Add feature X",
"author_id": "550e8400-e29b-41d4-a716-446655440001"
}'
```
Ответ: назначены 2 ревьювера из команды (не включая автора)


### Получение статистики
```bash
curl http://localhost:8080/stats
```

Ответ:
```bash
{
"user_stats": [...],
"pr_stats": {...},
"total_users": 3,
"total_teams": 1,
"active_users": 3
}
```

## Архитектура

```bash
HTTP Request
↓
Handler (Gin) ← валидация, маршрутизация
↓
Service ← бизнес-логика, алгоритм назначения
↓
Repository ← работа с PostgreSQL
↓
Database (PostgreSQL)
```


**Алгоритм назначения ревьюверов:**
1. Получить активных участников команды автора (исключая автора)
2. Подсчитать количество открытых PR для каждого участника
3. Случайно выбрать 2 ревьюверов из участников с наименьшей нагрузкой

## Технологии

- **Backend**: Go 1.23, Gin Framework
- **Database**: PostgreSQL 16 (с автоматическими миграциями)
- **Containerization**: Docker, Docker Compose
- **Testing**: Go test (unit), k6 (load)
- **Code Quality**: golangci-lint

## Тестирование

### Unit тесты
```bash
go test -v ./internal/...
```

С покрытием
```bash
go test -v -coverprofile=coverage.out ./internal/...
go tool cover -html=coverage.out
```

### Нагрузочные тесты
Требует запущенный сервис
```bash
k6 run loadtest/load_test.js
```
Результаты: p95 < 300ms, error rate < 1. Более подробное описание результатов в LOAD_TEST_RESULTS.md.

### Линтер
```bash
golangci-lint run
```

## Структура проекта
```bash
pr-reviewer-service/
├── cmd/api/ # Точка входа
├── internal/
│ ├── config/ # Конфигурация и БД
│ ├── domain/ # Модели и валидация
│ ├── handler/ # HTTP handlers (Gin)
│ ├── middleware/ # AdminAuth middleware
│ ├── repository/ # Database layer
│ └── service/ # Бизнес-логика
├── migrations/ # SQL миграции (auto-apply)
├── loadtest/ # k6 нагрузочные тесты
├── .golangci.yml # Конфигурация линтера
├── docker-compose.yml # Docker setup
├── Dockerfile # Multistage build
├── Makefile # Dev команды
└── README.md
```

## Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `DATABASE_URL` | PostgreSQL connection string | postgres://postgres:postgres@localhost:5432/pr_service?sslmode=disable |
| `PORT` | HTTP server port | 8080 |
| `ADMIN_TOKEN` | Token для admin endpoints | admin-secret |
| `LOG_LEVEL` | Уровень логирования | info |

## Makefile команды
```bash
make docker-up # Запуск в Docker
make docker-down # Остановка
make test # Unit тесты
make lint # Линтер
make load-test # k6 тесты
```

## Производительность

**SLI (Service Level Indicators):**
- Latency: p95 < 300ms
- Availability: > 99.9%
- Throughput: > 20 RPS

**Оптимизации:**
- Prepared statements для PostgreSQL
- Индексы на FK (team_id, user_id, pr_id)
- Connection pooling
- Транзакции для атомарных операций

## Разработка

### Локальный запуск (без Docker)
Запустить PostgreSQL
```bash
docker-compose up -d postgres
```

Установить зависимости
```bash
go mod download
```

Настроить переменные окружения
```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/pr_service?sslmode=disable"
export PORT=8080
```

Запустить сервис
```bash
go run cmd/api/main.go
```
