# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/passnow-api ./cmd/api
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/passnow ./cmd/passnow

FROM alpine:3.20

RUN addgroup -S passnow && adduser -S -G passnow passnow
WORKDIR /app

COPY --from=builder /out/passnow-api /app/passnow-api
COPY --from=builder /out/passnow /app/passnow

USER passnow
EXPOSE 8080

ENTRYPOINT ["/app/passnow-api"]
