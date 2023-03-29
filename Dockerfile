FROM golang:1 AS builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -v -o ./out/github-prometheus .

FROM alpine:3.17
RUN apk add ca-certificates

COPY --from=builder /app/out/github-prometheus /app/github-prometheus
EXPOSE 2199
CMD ["/app/github-prometheus"]
