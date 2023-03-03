.PHONY: build
build:
	mkdir -p bin && export CGO_ENABLED=0 GOOS=linux && go build -o bin/app -ldflags '-s -w' ./core/main.go

.PHONY: swagger
swagger:
ifeq ($(shell uname -m),arm64)
	docker build -t horizon-swagger -f build/swagger/Dockerfile . --platform linux/arm64
else
	docker build -t horizon-swagger -f build/swagger/Dockerfile .
endif

.PHONY: swagger-run
swagger-run: swagger
	@echo "http://localhost:80"
	docker run --rm -p 80:8080 horizon-swagger

.PHONY: job
job:
ifeq ($(shell uname -m),arm64)
	docker build -t horizon-job -f build/job/Dockerfile . --platform linux/arm64
else
	docker build -t horizon-job -f build/job/Dockerfile .
endif

.PHONY: core
core:
ifeq ($(shell uname -m),arm64)
	docker build -t horizon-core -f build/core/Dockerfile . --platform linux/arm64
else
	docker build -t horizon-core -f build/core/Dockerfile .
endif


.PHONY: clean
clean:
	docker rmi -f horizon-swagger
	docker rmi -f horizon-job
	docker rmi -f horizon-core

.PHONY: lint
lint:
	golangci-lint run -v

.PHONY: ut
ut:
	sh .unit-test.sh
