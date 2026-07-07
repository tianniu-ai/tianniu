FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./tianniu/main.go

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata sqlite-libs

COPY --from=builder /server /app/server

EXPOSE 8080

ENV GIN_MODE=release

CMD ["./server"]