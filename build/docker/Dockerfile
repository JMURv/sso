FROM golang:1.23.1-alpine3.20 as builder

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags "-s -w -extldflags '-static'" -o ./main ./cmd/main.go
RUN apk add upx
RUN upx ./main

FROM scratch

COPY ./internal/repo/db/migration ./internal/repo/db/migration
COPY --from=builder /app/go.mod ./go.mod
COPY --from=builder /app/main ./

ENTRYPOINT ["/main"]