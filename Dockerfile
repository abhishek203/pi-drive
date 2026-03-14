FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /pidrive-server ./cmd/pidrive-server

FROM alpine:3.20

RUN apk add --no-cache fuse3 curl bash

# Install JuiceFS
RUN curl -sSL https://d.juicefs.com/install | sh -

COPY --from=builder /pidrive-server /usr/local/bin/pidrive-server
COPY migrations/ /app/migrations/

WORKDIR /app

EXPOSE 8080

CMD ["pidrive-server"]
