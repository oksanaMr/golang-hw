BIN_CALENDAR := "./bin/calendar"
BIN_CALENDAR_SCHEDULER := "./bin/calendar_scheduler"
BIN_CALENDAR_STORER := "./bin/calendar_storer"

DOCKER_IMG_CALENDAR := "calendar-api:latest"
DOCKER_IMG_SCHEDULER := "calendar-scheduler:latest"
DOCKER_IMG_STORER := "calendar-storer:latest"

DOCKER_COMPOSE := "./deployments/docker-compose.yaml"
DOCKER_COMPOSE_TEST := "./test/docker-compose.yaml"
DOCKERFILE_CALENDAR := "./deployments/Dockerfile.calendar"
DOCKERFILE_SCHEDULER := "./deployments/Dockerfile.scheduler"
DOCKERFILE_STORER := "./deployments/Dockerfile.storer"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build:
	go build -v -o $(BIN_CALENDAR) -ldflags "$(LDFLAGS)" ./cmd/calendar
	go build -v -o $(BIN_CALENDAR_SCHEDULER) -ldflags "$(LDFLAGS)" ./cmd/calendar_scheduler
	go build -v -o $(BIN_CALENDAR_STORER) -ldflags "$(LDFLAGS)" ./cmd/calendar_storer

run: build
	$(BIN_CALENDAR) -config ./configs/calendar_config.yaml

run-scheduler: build
	$(BIN_CALENDAR_SCHEDULER) -config ./configs/calendar_config.yaml

run-storer: build
	$(BIN_CALENDAR_STORER) -config ./configs/calendar_config.yaml

build-img:
	docker build --build-arg=LDFLAGS="$(LDFLAGS)" -t $(DOCKER_IMG_CALENDAR) -f $(DOCKERFILE_CALENDAR) .
	docker build --build-arg=LDFLAGS="$(LDFLAGS)" -t $(DOCKER_IMG_SCHEDULER) -f $(DOCKERFILE_SCHEDULER) .
	docker build --build-arg=LDFLAGS="$(LDFLAGS)" -t $(DOCKER_IMG_STORER) -f $(DOCKERFILE_STORER) .

build-img-calendar:
	docker build --build-arg=LDFLAGS="$(LDFLAGS)" -t $(DOCKER_IMG_CALENDAR) -f $(DOCKERFILE_CALENDAR) .

build-img-scheduler:
	docker build --build-arg=LDFLAGS="$(LDFLAGS)" -t $(DOCKER_IMG_SCHEDULER) -f $(DOCKERFILE_SCHEDULER) .

build-img-storer:
	docker build --build-arg=LDFLAGS="$(LDFLAGS)" -t $(DOCKER_IMG_STORER) -f $(DOCKERFILE_STORER) .

run-img-calendar: build-img-calendar
	docker run --rm -p 8888:8888 $(DOCKER_IMG_CALENDAR)

run-img-scheduler: build-img-scheduler
	docker run --rm $(DOCKER_IMG_SCHEDULER)

run-img-storer: build-img-storer
	docker run --rm $(DOCKER_IMG_STORER)

up:
	docker-compose -f $(DOCKER_COMPOSE) up -d

logs:
	docker-compose -f $(DOCKER_COMPOSE) logs -f

down:
	docker-compose -f $(DOCKER_COMPOSE) down -v

restart-api:
	docker-compose -f $(DOCKER_COMPOSE) restart calendar-api

restart-scheduler:
	docker-compose -f $(DOCKER_COMPOSE) restart calendar-scheduler

restart-storer:
	docker-compose -f $(DOCKER_COMPOSE) restart calendar-storer

clean:
	docker-compose -f $(DOCKER_COMPOSE) down -v
	docker system prune -f

version: build
	$(BIN_CALENDAR) version

test:
	go test -race ./internal/...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.64.8

lint: install-lint-deps
	golangci-lint run ./...

integration-tests:
	docker-compose -f test/docker-compose.test.yaml up --build --abort-on-container-exit --exit-code-from integration-tests
	docker-compose -f test/docker-compose.test.yaml down -v

.PHONY: build run run-scheduler run-storer build-img build-img-calendar build-img-scheduler build-img-storer \
	run-img-calendar run-img-scheduler run-img-storer up logs down restart-api restart-scheduler restart-storer \
	clean version test lint install-lint-deps integration-tests