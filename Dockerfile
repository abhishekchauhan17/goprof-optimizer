# ---- Stage 1: Build ----
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git build-base

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_DATE=unknown

RUN go build -ldflags "\
   -X 'github.com/yourname/goprof-optimizer/internal/version.Version=${VERSION}' \
   -X 'github.com/yourname/goprof-optimizer/internal/version.Commit=${COMMIT}' \
   -X 'github.com/yourname/goprof-optimizer/internal/version.BuildDate=${BUILD_DATE}' \
" -o goprof-optimizer ./cmd/profiler

# ---- Stage 2: Runtime ----
FROM alpine:3

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=builder /app/goprof-optimizer .
COPY config.example.yaml ./config.yaml

USER app

EXPOSE 8080

ENTRYPOINT ["./goprof-optimizer", "-config", "/app/config.yaml"]
