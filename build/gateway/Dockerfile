FROM openresty/openresty:alpine

ARG GROUP=netease
ARG USER=appops
ARG GROUP_ID=10001
ARG USER_ID=10001

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
        apk update && apk add bash perl curl && opm install zmartzone/lua-resty-openidc

RUN addgroup --gid $GROUP_ID $GROUP && adduser -h /home/$USER -u $USER_ID -G $GROUP -D $USER && \
    sed -i "s/.*user.*nobody.*/user $USER $GROUP;/" /usr/local/openresty/nginx/conf/nginx.conf && \
    chown -R $USER:$GROUP /var/run/ /usr/local/openresty

USER $USER
