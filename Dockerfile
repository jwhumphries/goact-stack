# =============================================================================
# Development Dockerfile
# Production builds are handled by Dagger (task build)
# =============================================================================

# Stage: gobase - Go toolchain with hot-reload tools
FROM golang:1.24-alpine AS gobase

RUN apk add --no-cache git
RUN go install github.com/air-verse/air@latest

# Stage: dev - Development environment
FROM ghcr.io/jwhumphries/frontend:latest AS dev

WORKDIR /app

# Copy Go toolchain from gobase
COPY --from=gobase /usr/local/go /usr/local/go
COPY --from=gobase /go/bin/air /usr/local/bin/air

# Environment setup
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/go"
ENV GOCACHE=/go-build-cache
ENV GOMODCACHE=/go/pkg/mod

# Copy development script
COPY scripts/develop.sh /develop.sh
RUN chmod +x /develop.sh

EXPOSE 8080 3000

CMD ["/develop.sh"]
