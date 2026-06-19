# Stage 1: Build the Go binary
FROM golang:1.26-alpine AS builder

# Install necessary build tools (git, etc.)
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod and sum files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source code
COPY . .

# Build the executable as a static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api/main.go

# Stage 2: Create a minimal image
FROM alpine:3.21

# Install CA certificates to allow HTTPS requests (e.g. to Elasticsearch/Stripe/etc)
# and tzdata for timezone support
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/main .

# Expose port 8000
EXPOSE 8000

# Command to run the executable
CMD ["./main"]

