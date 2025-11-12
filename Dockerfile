FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./cmd/http/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/app .

CMD ["./app"]
