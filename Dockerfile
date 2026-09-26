FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /build/my-app ./cmd/main.go

FROM alpine:3.20

RUN adduser -D -u 10001 appuser

WORKDIR /app

COPY --from=builder /build/my-app /app/my-app

USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/my-app"]