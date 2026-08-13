COVERAGE_MINIMUM := 80
STRESS_PEAK ?= 1000
WORKER_REPLICAS ?= 1
COMPOSE := docker compose
STACK := $(COMPOSE) -f docker-compose.yml
INTERFACE_STACK := $(STACK) -f docker-compose.interfaces.yml
# Interface containers are intentionally outside the ephemeral load definition.
LOAD_STACK := COMPOSE_IGNORE_ORPHANS=true $(STACK) -f docker-compose.load.yml

.PHONY: build up rebuild up-interfaces down logs ps scale-workers swagger deps coverage load-smoke load-sustainable load-saturation load-stress

# Docker Compose helpers for the complete local stack.

build:
	$(STACK) build

up:
	$(STACK) up -d --scale worker=$(WORKER_REPLICAS)

rebuild:
	$(STACK) up --build -d --scale worker=$(WORKER_REPLICAS)

up-interfaces:
	$(INTERFACE_STACK) up -d --scale worker=$(WORKER_REPLICAS)

scale-workers:
	$(STACK) up -d --scale worker=$(WORKER_REPLICAS) worker

down:
	$(INTERFACE_STACK) down

logs:
	$(INTERFACE_STACK) logs -f

ps:
	$(INTERFACE_STACK) ps

swagger:
	go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g main.go -d cmd/api,internal/core/error,internal/inbound/http/adapter,internal/inbound/http/dto -o docs

deps:
	go mod download

# coverage measures unit-testable application code without external drivers.
coverage:
	go test -coverprofile=coverage.out $$(go list ./... | grep -Ev '/(cmd|docs)(/|$$)|/pkg/(mongo|rabbitmq)$$|/internal/outbound/outbox/adapter$$')
	go tool cover -func=coverage.out
	go tool cover -func=coverage.out | awk -v minimum=$(COVERAGE_MINIMUM) '/^total:/ { value = $$3; sub("%", "", value); if (value + 0 < minimum) { printf "coverage %.1f%% is below %.1f%%\n", value, minimum > "/dev/stderr"; exit 1 } }'

load-smoke:
	$(STACK) up -d --scale worker=$(WORKER_REPLICAS) worker
	$(LOAD_STACK) run --rm k6 run -e K6_PROFILE=smoke -e WORKER_REPLICAS=$(WORKER_REPLICAS) /scripts/orders.js

load-sustainable:
	$(STACK) up -d --scale worker=$(WORKER_REPLICAS) worker
	$(LOAD_STACK) run --rm k6 run -e K6_PROFILE=sustainable -e WORKER_REPLICAS=$(WORKER_REPLICAS) /scripts/orders.js

load-saturation:
	$(STACK) up -d --scale worker=$(WORKER_REPLICAS) worker
	$(LOAD_STACK) run --rm k6 run -e K6_PROFILE=saturation -e WORKER_REPLICAS=$(WORKER_REPLICAS) /scripts/orders.js

load-stress:
	$(STACK) up -d --scale worker=$(WORKER_REPLICAS) worker
	$(LOAD_STACK) run --rm k6 run -e K6_PROFILE=stress -e K6_STRESS_PEAK=$(STRESS_PEAK) -e WORKER_REPLICAS=$(WORKER_REPLICAS) /scripts/orders.js
