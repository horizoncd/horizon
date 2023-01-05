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

COPY ./openapi /openapi
COPY ./build/swagger/build.sh /openapi

RUN cp /openapi/build.sh /openapi/v1/restful/build.sh && cd /openapi/v1/restful && bash build.sh && \
    cp /openapi/build.sh /openapi/v1/front/build.sh && cd /openapi/v1/front && bash build.sh && \
    cp /openapi/v1/restful/swagger.json /openapi/restful-v1.0.0.json  && \
    cp /openapi/v1/front/swagger.json /openapi/front-v1.0.0.json && \
    cp /openapi/build.sh /openapi/v2/restful/build.sh && cd /openapi/v2/restful && bash build.sh && \
    cp /openapi/build.sh /openapi/v2/front/build.sh && cd /openapi/v2/front && bash build.sh && \
    cp /openapi/v2/restful/swagger.json /openapi/restful-v2.0.0.json && \
    cp /openapi/v2/front/swagger.json /openapi/front-v2.0.0.json && \

RUN chown -R $USER:$GROUP /openapi

USER $USER

