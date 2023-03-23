.DEFAULT_GOAL := help
SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

INSTALL_BIN_DIR = /usr/local/bin

# Declare all targets as phony targets
.PHONY: all build swagger swagger-run job core clean lint ut imports help 

all: tidy gen format lint cover build

## build: Build the project for horizon
build:
	@mkdir -p bin && export CGO_ENABLED=0 && go build -o bin/app -ldflags '-s -w' ./core/main.go

## swagger: Build the swagger
swagger:
ifeq ($(shell uname -m),arm64)
	@docker build -t horizon-swagger -f build/swagger/Dockerfile . --platform linux/arm64
else
	@docker build -t horizon-swagger -f build/swagger/Dockerfile .
endif

## Run a swagger server locally
swagger-run: swagger
	@echo "===========> Swagger is available at http://localhost:80"
	@docker run --rm -p 80:8080 horizon-swagger

## job: Build the job
job:
ifeq ($(shell uname -m),arm64)
	@docker build -t horizon-job -f build/job/Dockerfile . --platform linux/arm64
else
	@docker build -t horizon-job -f build/job/Dockerfile .
endif

## core: Build the core
core:
ifeq ($(shell uname -m),arm64)
	@docker build -t horizon-core -f build/core/Dockerfile . --platform linux/arm64
else
	@docker build -t horizon-core -f build/core/Dockerfile .
endif

## clean: Clean the project and remove the docker images
clean:
	@echo "===========> Cleaning all build output"
	@docker rmi -f horizon-swagger
	@docker rmi -f horizon-job
	@docker rmi -f horizon-core

## lint: Run the golangci-lint
lint:
	@echo "===========> Linting the code"
	@golangci-lint run --verbose

## ut: Run the unit tests
ut:
	@sh .unit-test.sh

## imports: task to automatically handle import packages in Go files using goimports tool
imports:
	@goimports -l -w $(SRC)

## k3s-uninstall: Unnstall k3s
k3s-uninstall:
	(cd $(INSTALL_BIN_DIR) && ./k3s-killall.sh) || true
	(cd $(INSTALL_BIN_DIR) && ./k3s-uninstall.sh) || true

## k3s-install: Install k3s
k3s-install:
	@echo "===========> Installing k3s"
	cd scripts && chmod +x install.sh && ./install.sh --k3s

## help: Display help information
help: Makefile
	@echo ""
	@echo "Usage:"
	@echo ""
	@echo "  make [comment]"
	@echo ""
	@echo "Comments:"
	@echo ""
	@awk -F ':|##' '/^[^\.%\t][^\t]*:.*##/{printf "  \033[36m%-20s\033[0m %s\n", $$1, $$NF}' $(MAKEFILE_LIST) | sort
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'