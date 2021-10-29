
generate:
	protoc --proto_path=api/proto --go_out=. --go-grpc_out=. --grpc-gateway_out=. --validate_out="lang=go:." api/proto/EventService.proto

lint:
	golangci-lint run ./...

up:
	docker-compose up -d

rebuild:
	docker-compose up --build -d

down:
	docker-compose down

build:
	go build -v -o ./bin/calendar-api ./cmd/calendar
	go build -v -o ./bin/scheduler ./cmd/scheduler
	go build -v -o ./bin/sender ./cmd/sender

run:
	./bin/calendar -config ./configs/calendar_config.yml

test:
	go test -count=1 -v ./internal/...

teardown-integration-tests:
	docker-compose -f docker-compose.test.yml down

setup-integration-tests: teardown-integration-tests
	docker-compose -f docker-compose.test.yml up --build -d
	sleep 3

start-integration-tests: setup-integration-tests
	GRPC_PORT="8889" GRPC_HOST="127.0.0.1" PGSQL_DSN="host=localhost port=5432 user=user password=secret dbname=calendar_tests sslmode=disable" go test -v ./tests/integration/ -tags integration -count=1
