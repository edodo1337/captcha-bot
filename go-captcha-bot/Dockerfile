FROM golang:1.20.2-alpine AS builder
RUN apk add -U tzdata
RUN apk --update add ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY config.yaml ./
COPY . .
RUN CGO_ENABLED=0 go build -ldflags '-w -s -extldflags "-static"' -o main ./cmd/

FROM scratch
WORKDIR /app
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /app/main .
COPY --from=builder /app/config.yaml .
CMD ["./main"]
