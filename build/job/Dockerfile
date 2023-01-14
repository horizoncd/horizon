FROM golang:1.15.3 AS builder
COPY . /horizon

WORKDIR /horizon

RUN CGO_ENABLED=0 GOOS=linux go build -o bin/app -ldflags '-s -w' ./job/main.go

FROM alpine/alpine:3.9 AS runtime

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk update && apk add bash curl git && \
    apk add tzdata && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo "Asia/Shanghai" > /etc/timezone

ARG GROUP=netease
ARG USER=appops
ARG GROUP_ID=10001
ARG USER_ID=10001

COPY --from=builder /horizon/bin/app /usr/local/bin/app

RUN addgroup --gid $GROUP_ID $GROUP && adduser -h /home/$USER -u $USER_ID -G $GROUP -D $USER && \
    chown -R $USER:$GROUP /usr/local/bin/app


RUN  echo "hosts: files dns" > /etc/nsswitch.conf

USER $USER

ENTRYPOINT ["/usr/local/bin/app"]
