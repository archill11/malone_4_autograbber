up:
	docker-compose up
.PHONY:up

runs:
	go build -o bin/myapp ./cmd/main.go && ./bin/myapp
.PHONY:runs

runss:
	go run ./cmd/main.go
.PHONY:runss

fmt:
	go fmt ./...
.PHONY:fmt

lint: fmt
	golint ./...
.PHONY:lint

vet: fmt
	go vet ./...
.PHONY:vet