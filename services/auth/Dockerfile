# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy shared module first (for layer caching)
COPY shared/go.mod shared/go.sum* ./shared/
RUN cd shared && go mod download

# Copy auth service module
COPY services/auth/go.mod services/auth/go.sum* ./services/auth/
RUN cd services/auth && go mod download

# Copy source code
COPY shared/ ./shared/
COPY services/auth/ ./services/auth/
COPY go.work ./

# Build
WORKDIR /app/services/auth
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/server cmd/server/main.go

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 -S dealance && \
    adduser -u 1000 -S dealance -G dealance

WORKDIR /app

# Copy binary
COPY --from=builder /app/server .

# Copy migrations (needed for migrate command)
COPY --from=builder /app/services/auth/migrations ./migrations/

# Run as non-root
USER dealance

EXPOSE 8080

ENTRYPOINT ["./server"]
