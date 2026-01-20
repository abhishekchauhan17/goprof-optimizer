APP_NAME = goprof-optimizer
CMD = ./cmd/profiler

VERSION ?= dev
COMMIT  := $(shell git rev-parse --short HEAD || echo "none")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS = -X 'github.com/AbhishekChauhan17/goprof-optimizer/internal/version.Version=$(VERSION)' \
          -X 'github.com/AbhishekChauhan17/goprof-optimizer/internal/version.Commit=$(COMMIT)' \
          -X 'github.com/AbhishekChauhan17/goprof-optimizer/internal/version.BuildDate=$(BUILD_DATE)'

all: build

build:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(APP_NAME) $(CMD)

run:
	go run -ldflags "$(LDFLAGS)" $(CMD) -config=config.example.yaml

test:
	go test ./... -race -count=1

bench:
	go test -bench=. -benchmem ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/

docker-build:
	docker build -t $(APP_NAME):latest .

docker-run:
	docker run -p 8080:8080 $(APP_NAME):latest
