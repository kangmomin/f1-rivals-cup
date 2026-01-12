# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install migrate tool
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary and migrate tool from builder
COPY --from=builder /app/server .
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY --from=builder /app/db/migrations ./db/migrations

# Create start script
RUN echo '#!/bin/sh' > /app/start.sh && \
    echo 'echo "Running database migrations..."' >> /app/start.sh && \
    echo 'migrate -path /app/db/migrations -database "$DATABASE_URL" up' >> /app/start.sh && \
    echo 'echo "Starting server..."' >> /app/start.sh && \
    echo 'exec /app/server' >> /app/start.sh && \
    chmod +x /app/start.sh

EXPOSE 8080

CMD ["/app/start.sh"]
