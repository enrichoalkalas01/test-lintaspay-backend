.PHONY: dev build run tidy swagger test migrate migrate-down

dev:
	air

build:
	go build -o ./tmp/main.exe ./cmd/api

run: build
	./tmp/main.exe

tidy:
	go mod tidy

swagger:
	swag init -g cmd/api/server.go -o docs --parseDependency --parseInternal

test:
	go test ./...

migrate:
	go run ./cmd/migrate

migrate-down:
	go run ./cmd/migrate -down
