FROM golang:1.15.3 AS builder
COPY .. /horizon

WORKDIR /horizon

RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o bin/app -ldflags '-s -w' ./core/cmd/main.go

FROM alpine:3.9 AS runtime

COPY --from=builder /horizon/bin/app /usr/local/bin/app
EXPOSE 8080/tcp
CMD ["/usr/local/bin/app"]
