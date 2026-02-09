FROM golang:1.22-alpine AS builder
WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go test ./...
RUN CGO_ENABLED=1 go test -race ./...
RUN go build -trimpath -o /lab-server ./cmd/lab-server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /lab-server .
EXPOSE 8080
CMD ["./lab-server", "-port", "8080"]
