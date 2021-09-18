FROM swaggerapi/swagger-ui:latest

ARG GROUP=netease
ARG USER=appops
ARG GROUP_ID=10001
ARG USER_ID=10001

RUN addgroup --gid $GROUP_ID $GROUP && adduser -h /home/$USER -u $USER_ID -G $GROUP -D $USER && \
    chown -R $USER:$GROUP /var/run/ /usr/ && \
    sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && apk update && apk add bash && \
    curl -#sSL -o /usr/bin/jq https://music-cloudnative.nos-jd.163yun.com/binary/jq-linux64 && \
    curl -#sSL -o /usr/bin/yaml2json https://music-cloudnative.nos-jd.163yun.com/binary/yaml2json_linux_amd64 && \
    chmod +x /usr/bin/jq && chmod +x /usr/bin/yaml2json

COPY ./openapi/restful /restful
COPY ./build/swagger/build.sh /restful

RUN cd /restful && bash build.sh

RUN chown -R $USER:$GROUP /restful

USER $USER

