FROM golang:1.15.3 AS builder
COPY . /horizon

WORKDIR /horizon

RUN CGO_ENABLED=0 GOOS=linux go build -o bin/app -ldflags '-s -w' ./core/main.go

FROM alpine:3.9 AS runtime

ARG GROUP=netease
ARG USER=appops
ARG GROUP_ID=10001
ARG USER_ID=10001

COPY --from=builder /horizon/bin/app /usr/local/bin/app

RUN addgroup --gid $GROUP_ID $GROUP && adduser -h /home/$USER -u $USER_ID -G $GROUP -D $USER && \
    chown -R $USER:$GROUP /usr/local/bin/app

USER $USER

CMD ["/usr/local/bin/app"]
