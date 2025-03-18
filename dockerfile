FROM golang:1.23 AS builder
WORKDIR /app
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -o app main.go

FROM alpine
WORKDIR /
COPY --from=builder /app/app /app
RUN chmod +x /app
CMD ["/app"]
