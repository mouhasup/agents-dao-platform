# Build stage
FROM golang:alpine AS builder

WORKDIR /app

# Install git and certificates required for downloading dependencies
RUN apk add --no-cache git ca-certificates

# ── Cache layer: copy only module files first so `go mod download` is cached
#    and only re-runs when go.mod / go.sum change, NOT on every source edit.
COPY go.mod go.sum ./
RUN go mod download

# ── Copy the rest of the source and build
COPY . .
RUN go mod tidy

# Build a statically linked binary for Linux without CGO
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server main.go

# Production runtime stage
FROM alpine:latest

# Install timezone data and root certificates for outbound HTTPS calls (Google APIs, GenAI)
RUN apk add --no-cache tzdata ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Copy credentials and environment files if present
COPY credentials.json* service-account.json* .env* ./

EXPOSE 8080

CMD ["./server"]
