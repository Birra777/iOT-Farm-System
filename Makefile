.PHONY: up down run-simulator run-processor run-anomaly run-api run-predictor run-dashboard migrate logs help

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

migrate:
	go run ./cmd/migrate

run-simulator:
	go run ./cmd/simulator

run-processor:
	go run ./cmd/processor

run-anomaly:
	go run ./cmd/anomaly

run-predictor:
	go run ./cmd/predictor

run-api:
	go run ./cmd/api

run-dashboard:
	cd dashboard && npm run dev

test:
	go test ./...

test-integration:
	go test -tags=integration ./...

help:
	@echo "Available targets:"
	@echo "  up              - Start Docker services (Kafka, PostgreSQL)"
	@echo "  down            - Stop Docker services"
	@echo "  logs            - Tail Docker service logs"
	@echo "  migrate         - Run database migrations"
	@echo "  run-simulator   - Start sensor simulator"
	@echo "  run-processor   - Start stream processor"
	@echo "  run-anomaly     - Start anomaly detector"
	@echo "  run-predictor   - Start predictive alert service"
	@echo "  run-api         - Start REST API server"
	@echo "  run-dashboard   - Start React dashboard dev server (http://localhost:5173)"
	@echo "  test            - Run unit tests"
	@echo "  test-integration - Run integration tests"
