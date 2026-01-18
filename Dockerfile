# =============================================================================
# Stage: gobase - Go toolchain with hot-reload tools
# =============================================================================
FROM golang:1.24-alpine AS gobase

RUN apk add --no-cache git

# Install Air for hot-reloading
RUN go install github.com/air-verse/air@latest

# =============================================================================
# Stage: dev - Development environment
# =============================================================================
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

# =============================================================================
# Stage: build - Production build
# =============================================================================
FROM golang:1.24-alpine AS build

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy frontend build (assumes it's been built externally)
COPY internal/static/dist internal/static/dist

# Copy source code
COPY . .

# Build arguments for version injection
ARG VERSION=dev

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X goact-stack/version.Tag=${VERSION}" \
    -o /goact-stack ./cmd/goact-stack

# =============================================================================
# Stage: production - Minimal production image
# =============================================================================
FROM alpine:3.21 AS production

RUN apk add --no-cache tzdata ca-certificates

# Create nonroot user
RUN echo 'nonroot:x:10001:10001:NonRoot User:/:/sbin/nologin' >> /etc/passwd && \
    chmod 0600 /etc/shadow

ENV TZ=America/New_York
ENV GOACT_PORT=:8080

COPY --from=build /goact-stack /usr/local/bin/goact-stack

RUN chown -R 10001:10001 /usr/local/bin/goact-stack

USER 10001

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/goact-stack"]
