FROM golang:1.25-alpine3.21

WORKDIR /app

# Install air
RUN go install github.com/air-verse/air@latest

# Copy go.mod early to cache deps
COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["air", "-c", ".air.toml"]
