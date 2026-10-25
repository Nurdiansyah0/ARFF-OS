# Stage 1: Build static Go binary
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install git, ca-certificates, and tzdata
RUN apk update && apk add --no-cache git ca-certificates tzdata

# Copy module files and download dependencies
COPY go.mod ./
RUN go mod download || true

# Copy source code
COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/server ./cmd/api

# Stage 2: Minimal runtime image
FROM alpine:3.20

WORKDIR /app

# Copy SSL certificates & timezone data from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/server /app/server

# Run as non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

ENV PORT=8080
ENV APP_ENV=production

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:${PORT}/health || exit 1

CMD ["/app/server"]
