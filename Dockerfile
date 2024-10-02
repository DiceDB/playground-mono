# Stage 1: Build the Go application
FROM golang:1.23 AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download -x
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o playground-mono .

# Stage 2: Build the final container with the Go binary
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=go-builder /app/playground-mono .
EXPOSE 8080
CMD ["./playground-mono"]