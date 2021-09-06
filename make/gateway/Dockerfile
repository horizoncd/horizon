FROM openresty/openresty:alpine

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
        apk update && apk add bash perl curl

RUN opm install zmartzone/lua-resty-openidc

RUN addgroup --gid 10001 netease && adduser -h /home/appops -u 10001 -G netease -D appops && \
    sed -i 's/.*user.*nobody.*/user appops netease;/' /usr/local/openresty/nginx/conf/nginx.conf && \
    chown -R appops:netease /var/run/ && chown -R appops:netease /usr/local/openresty

USER appops
