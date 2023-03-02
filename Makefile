# Set the default target to build all
.DEFAULT_GOAL := all

# Declare all targets as phony targets
.PHONY: all build swagger swagger-run job core clean lint ut help

# Build all targets
all: tidy gen format lint cover build

## build: Build the project for horizon
build:
	mkdir -p bin && export CGO_ENABLED=0 && go build -o bin/app -ldflags '-s -w' ./core/main.go

## format: Gofmt (reformat) package sources (exclude vendor dir if existed).
.PHONY: format
format: tools.verify.golines tools.verify.goimports
	@echo "===========> Formating codes"
	@$(FIND) -type f -name '*.go' | $(XARGS) gofmt -s -w
	@$(FIND) -type f -name '*.go' | $(XARGS) goimports -w -local $(ROOT_PACKAGE)
	@$(FIND) -type f -name '*.go' | $(XARGS) golines -w --max-len=120 --reformat-tags --shorten-comments --ignore-generated .
	@$(GO) mod edit -fmt

## swagger: Build the swagger
swagger:
ifeq ($(shell uname -m),arm64)
	docker build -t horizon-swagger -f build/swagger/Dockerfile . --platform linux/arm64
else
	docker build -t horizon-swagger -f build/swagger/Dockerfile .
endif

## swagger-run: Run the swagger
swagger-run: swagger
	@echo "Swagger is available at http://localhost:80"
	docker run --rm -p 80:8080 horizon-swagger

## job: Build the job
job:
ifeq ($(shell uname -m),arm64)
	docker build -t horizon-job -f build/job/Dockerfile . --platform linux/arm64
else
	docker build -t horizon-job -f build/job/Dockerfile .
endif

## core: Build the core
core:
ifeq ($(shell uname -m),arm64)
	docker build -t horizon-core -f build/core/Dockerfile . --platform linux/arm64
else
	docker build -t horizon-core -f build/core/Dockerfile .
endif

## clean: Clean the project and remove the docker images
clean:
	@echo "===========> Cleaning all build output"
	@docker rmi -f horizon-swagger
	@docker rmi -f horizon-job
	@docker rmi -f horizon-core

## lint: Run the linter
lint:
	@golangci-lint run -v

## ut: Run the unit tests
ut:
	sh .unit-test.sh

## format: Run go fmt against code.
fmt:	
	@go fmt ./...

## help: Display help information
help: Makefile
	@echo "Usage:"
	@echo ""
	@echo "  make [target]"
	@echo ""
#	@echo "Targets:"
	@echo ""
	@awk -F ':|##' '/^[^\.%\t][^\t]*:.*##/{printf "  \033[36m%-20s\033[0m %s\n", $$1, $$NF}' $(MAKEFILE_LIST) | sort
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
.PHONY: help
