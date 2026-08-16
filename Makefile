.PHONY: build test fmt run migrate migrate-status docker-dev

build:
	go build ./...

test:
	go test ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

run:
	go run ./cmd/passnow serve

migrate:
	go run ./cmd/passnow migrate

migrate-status:
	go run ./cmd/passnow migrate status

docker-dev:
	docker compose -f docker-compose.dev.yml up --build
