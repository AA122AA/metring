build:
	@go build -o bin/server/server ./cmd/server/main.go

run: build
	@./bin/server/server
