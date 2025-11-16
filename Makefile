.PHONY: run build test docker-up docker-down migrate-up migrate-down clean

# Локальный запуск
run:
	go run cmd/api/main.go

# Сборка бинарника
build:
	go build -o bin/api cmd/api/main.go

# Тесты
test:
	go test -v ./...

# Запуск через Docker
docker-up:
	docker-compose up --build

# Остановка Docker
docker-down:
	docker-compose down

# Очистка volumes
clean:
	docker-compose down -v

# Создание новой миграции
create-migration:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

# Применить миграции
migrate-up:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/pr_service?sslmode=disable" up

# Откатить миграции
migrate-down:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/pr_service?sslmode=disable" down

# Форматирование кода
fmt:
	go fmt ./...

# Линтинг
lint:
	golangci-lint run

