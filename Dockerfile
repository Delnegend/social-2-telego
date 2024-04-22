from golang:1.22.2-alpine as builder

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -o main .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/main .

CMD ["./main"]