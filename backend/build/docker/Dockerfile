FROM golang:1.23.1-alpine3.20 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags "-s -w -extldflags '-static'" -o ./main ./cmd/main.go
RUN apk add upx
RUN upx ./main

FROM scratch
ENV MIGRATIONS_PATH=.

COPY ./internal/repo/db/migration ./
COPY --from=builder /app/main ./

ENTRYPOINT ["/main"]