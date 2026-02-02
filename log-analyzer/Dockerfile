# Build stage for frontend
FROM node:18-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# Build stage for backend
FROM golang:1.21-alpine AS backend-builder
WORKDIR /app
# Copy go.mod first for better caching
COPY go.mod ./
# Download dependencies
RUN go mod download
# Copy source code
COPY . .
# Generate go.sum after copying source code
RUN go mod tidy
# Build binaries
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/cli ./backend/cmd/cli
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api ./backend/cmd/api

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

# Copy binaries
COPY --from=backend-builder /app/bin/cli /app/cli
COPY --from=backend-builder /app/bin/api /app/api

# Copy frontend build
COPY --from=frontend-builder /app/frontend/dist /app/frontend/dist

# Copy config
COPY config /app/config

# Create logs directory
RUN mkdir -p /var/log

EXPOSE 8080

# Default to API server, can be overridden
CMD ["./api"]
