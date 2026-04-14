FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./cmd/http/main.go

FROM alpine:latest

WORKDIR /root/

RUN mkdir -p /etc/app

COPY --from=builder /app/app .

ENTRYPOINT ["./app"]
CMD ["-config /etc/app/local.yaml"]
