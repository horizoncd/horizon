# Build all by default, even if it's not first
.DEFAULT_GOAL := help

# ==============================================================================
# Build options
JOB_NAME := horizon-job
SWAGGER_NAME := horizon-swagger
CORE_NAME := horizon-core

SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")
# ==============================================================================
# Includes


# ==============================================================================
# Targets

## all: Follow the steps required for development in sequence
.PHONY: all
all: tidy gen lint build

## build: Build the project for horizon
.PHONY: build
build:
	@mkdir -p bin && export CGO_ENABLED=0 && go build -o bin/app -ldflags '-s -w' ./core/main.go

## swagger: Build the swagger
.PHONY: swagger
swagger:
ifeq ($(shell uname -m),arm64)
	@docker build -t $(SWAGGER_NAME) -f build/swagger/Dockerfile . --platform linux/arm64
else
	@docker build -t $(SWAGGER_NAME) -f build/swagger/Dockerfile .
endif

## s Run a swagger server locally
.PHONY: swagger-run
swagger-run: swagger
	@echo "===========> Swagger is available at http://localhost:80"
	@docker run --rm -p 80:8080 $(SWAGGER_NAME)

## job: Build the job
.PHONY: job
job:
ifeq ($(shell uname -m),arm64)
	@docker build -t $(JOB_NAME) -f build/job/Dockerfile . --platform linux/arm64
else
	@docker build -t $(JOB_NAME) -f build/job/Dockerfile .
endif

## core: Build the core
.PHONY: core
core:
ifeq ($(shell uname -m),arm64)
	@docker build -t $(CORE_NAME) -f build/core/Dockerfile . --platform linux/arm64
else
	@docker build -t $(CORE_NAME) -f build/core/Dockerfile .
endif

## clean: Clean the project and remove the docker images
.PHONY: clean
clean:
	@echo "===========> Cleaning all build output"
	@docker rmi -f $(SWAGGER_NAME)
	@docker rmi -f $(JOB_NAME)
	@docker rmi -f $(CORE_NAME)

## lint: Run the golangci-lint
.PHONY: lint
lint:
	@echo "===========> Linting the code"
	@golangci-lint run --verbose

## gen: Generate all necessary files, such as mock code.
gen:
	@echo "===========> Installing codegen"
	@go generate ./...

## ut: Run the unit tests
.PHONY: ut
ut:
	@sh .unit-test.sh

## imports: task to automatically handle import packages in Go files using goimports tool
.PHONY: imports
imports:
	@goimports -l -w $(SRC)

## help: Display help information
help: Makefile
	@echo ""
	@echo "Usage:" "\n"
	@echo "  make [target]" "\n"
	@echo "Targets:" "\n" ""
	@awk -F ':|##' '/^[^\.%\t][^\t]*:.*##/{printf "  \033[36m%-20s\033[0m %s\n", $$1, $$NF}' $(MAKEFILE_LIST) | sort
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'