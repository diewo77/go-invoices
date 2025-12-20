# syntax=docker/dockerfile:1
FROM node:20-alpine AS node_deps
WORKDIR /app/billing-app
COPY billing-app/package.json billing-app/package-lock.json* ./
RUN if [ -f package.json ]; then npm install --no-audit --no-fund; fi

FROM golang:1.25.5 AS builder
WORKDIR /app/billing-app
# Module files + local module sources to satisfy replace directives
COPY billing-app/go.mod billing-app/go.sum ./
COPY auth /app/auth
COPY httpx /app/httpx
COPY i18n /app/i18n
COPY pdf /app/pdf
COPY validation /app/validation
COPY view /app/view
RUN go mod download
COPY billing-app/ .
# Copy installed node modules for Tailwind build if present
COPY --from=node_deps /app/billing-app/node_modules ./node_modules
RUN chmod +x build-assets.sh || true
RUN if [ -f package.json ]; then ./build-assets.sh || echo "Skipping CSS build"; fi
RUN CGO_ENABLED=0 GOOS=linux go build -o billing-app ./cmd/server

# Development stage: includes toolchain + reflex auto-reloader
FROM golang:1.25.5 AS dev
WORKDIR /app/billing-app
COPY billing-app/go.mod billing-app/go.sum ./
COPY auth /app/auth
COPY httpx /app/httpx
COPY i18n /app/i18n
COPY pdf /app/pdf
COPY validation /app/validation
COPY view /app/view
RUN go mod download && go install github.com/cespare/reflex@latest
COPY --from=node_deps /app/billing-app/node_modules ./node_modules
COPY billing-app/ .
## Install node/npm so we can run tailwind in watch mode inside dev container (Debian base)
RUN apt-get update \
    && apt-get install -y --no-install-recommends nodejs npm inotify-tools \
    && rm -rf /var/lib/apt/lists/*
## Install golang-migrate CLI for running migrations from inside the dev container
RUN wget -qO- https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz \
    | tar xz -C /usr/local/bin && chmod +x /usr/local/bin/migrate
RUN chmod +x dev.sh build-assets.sh || true
ENV DEV=1
EXPOSE 8080 5173
# Dev entrypoint handled via docker-compose command (dev.sh available if needed)

FROM alpine:3.19 AS final
WORKDIR /app
RUN adduser -D -g '' appuser
COPY --from=builder /app/billing-app/billing-app ./
COPY --from=builder /app/billing-app/static ./static
COPY --from=builder /app/billing-app/migrations ./migrations
COPY templates ./templates
USER appuser
EXPOSE 8080
ENV GIN_MODE=release
CMD ["./billing-app"]
