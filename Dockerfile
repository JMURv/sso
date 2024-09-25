FROM golang:1.23.1-alpine3.20 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main ./cmd/main.go

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/main ./

EXPOSE 50050

CMD ["./main"]
