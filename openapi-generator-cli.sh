# ref: https://github.com/OpenAPITools/openapi-generator#16---docker

# 1. group
docker run --rm -v "$PWD:/local" openapitools/openapi-generator-cli generate \
    -i /local/openapi/group.yaml \
    -g go-gin-server \
    -c /local/openapi/group-config.yaml \
    -o /local/http

# 2. service
docker run --rm -v "$PWD:/local" openapitools/openapi-generator-cli generate \
    -i /local/openapi/service.yaml \
    -g go-gin-server \
    -c /local/openapi/service-config.yaml \
    -o /local/http

# 3. login
docker run --rm -v "$PWD:/local" openapitools/openapi-generator-cli generate \
    -i /local/openapi/login.yaml \
    -g go-gin-server \
    -c /local/openapi/login-config.yaml \
    -o /local/http