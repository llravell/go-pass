include .env
export

LOCAL_BIN ?= $(CURDIR)/bin
GO_PASS_DIR = $(HOME)/.go_pass

.PHONY: lint
lint: ### run linter
	golangci-lint run ./...

.PHONY: test
test: ### run tests
	go test ./...

.PHONY: server
server: ### run grpc server
	go run ./cmd/server/main.go

.PHONY: server-migrations
server-migrations: ### up server migrations
	go run ./cmd/server-migrations/main.go

.PHONY: reqs
reqs: ### install binary deps to bin/
	GOBIN=$(LOCAL_BIN) go install go.uber.org/mock/mockgen@latest
	GOBIN=$(LOCAL_BIN) go install github.com/pressly/goose/v3/cmd/goose@latest
