FROM golang:latest AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the playground binary
RUN go build -o vuego-playground ./cmd/vuego-playground

FROM golang:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/vuego-playground .

EXPOSE 8080

ENTRYPOINT ["./vuego-playground"]
