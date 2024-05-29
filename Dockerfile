FROM golang:1.22.3-alpine as builder

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -o main .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/main .

CMD ["./main"]