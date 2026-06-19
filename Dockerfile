# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/subscriptions-service .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata \
	&& adduser -D -H -u 10001 appuser

WORKDIR /app
COPY --from=builder /out/subscriptions-service /app/subscriptions-service

ENV PORT=8080
EXPOSE 8080

USER appuser
CMD ["/app/subscriptions-service"]
