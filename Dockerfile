# --- Build stage ---
FROM golang:1.25-alpine3.21 AS build

WORKDIR /app

# Copy go.mod early for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary
RUN go build -o app ./cmd/rip

# --- Run stage ---
FROM alpine:3.21

WORKDIR /app

# Copy binary from builder
COPY --from=build /app/app .
COPY config.toml cert.crt cert.key ./

# Expose port (если нужно)
EXPOSE 8000

# Run app
CMD ["./app"]
