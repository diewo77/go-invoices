# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go.work and all module files
COPY go.work go.work.sum ./

# Copy all modules
COPY auth/ ./auth/
COPY httpx/ ./httpx/
COPY i18n/ ./i18n/
COPY validation/ ./validation/
COPY view/ ./view/
COPY go-gate/ ./go-gate/
COPY go-pdf/ ./go-pdf/
COPY go-invoices/ ./go-invoices/

# Download dependencies
WORKDIR /app/go-invoices
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/server ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Copy templates and static files
COPY --from=builder /app/go-invoices/templates ./templates

# Expose port
EXPOSE 8080

# Run
CMD ["./server"]
