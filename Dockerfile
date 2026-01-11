FROM golang:1.25-alpine3.23 AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o ssoapp ./cmd/isso/main.go

FROM alpine:3.23

RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/ssoapp .
COPY --from=builder /app/static ./static
CMD ["./ssoapp"]