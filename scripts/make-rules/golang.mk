# Copyright 2023 The Horizoncd Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# ==============================================================================
# define the default goal
#

GO := go
# !I have only tested gvm since 1.18. You are advised to use the gvm switchover version of the tools toolkit
GO_SUPPORTED_VERSIONS ?= |1.15|1.16|1.17|1.18|1.19|1.20|

ifeq ($(ROOT_PACKAGE),)
	$(error the variable ROOT_PACKAGE must be set prior to including golang.mk, -> /Makefile)
endif

GOPATH ?= $(shell go env GOPATH)
ifeq ($(origin GOBIN), undefined)
	GOBIN := $(GOPATH)/bin
endif

# Unit tests exclude directories
EXCLUDE_TESTS= $(ROOT_DIR)/scripts $(ROOT_DIR)/image $(ROOT_DIR)/test

## go.build: Build binaries
.PHONY: go.build
go.build:
	@echo "$(shell go version)"
	@echo "===========> Building binary $(BUILDAPP) *[Git Info]: $(VERSION)-$(GIT_TAG)-$(GIT_COMMIT)"
	@export CGO_ENABLED=0 && go build -o $(BUILDAPP) -ldflags '-s -w' $(BUILDFILE)

## swagger-run: Run a swagger server locally
.PHONY: go.swagger-run
go.swagger-run: image.swagger
	@echo "===========> Swagger is available at http://localhost:80"
	@docker run --rm -p 80:8080 horizon-swagger

## go.test: Run unit test
go.test:
	@echo "===========> Run unit test"
	@$(GO) test ./...

## go.test.junit-report: Run unit test
.PHONY: go.test.junit-report
go.test.junit-report: tools.verify.go-junit-report
	@echo "===========> Run unit test > $(OUTPUT_DIR)/report.xml"
	@$(GO) test -v -coverprofile=$(OUTPUT_DIR)/coverage.out 2>&1 ./... | $(TOOLS_DIR)/go-junit-report -set-exit-code > $(OUTPUT_DIR)/report.xml
	@sed -i '/mock_.*.go/d' $(OUTPUT_DIR)/coverage.out
	@echo "===========> Test coverage of Go code is reported to $(OUTPUT_DIR)/coverage.html by generating HTML"
	@$(GO) tool cover -html=$(OUTPUT_DIR)/coverage.out -o $(OUTPUT_DIR)/coverage.html

# .PHONY: go.test.junit-report
# go.test.junit-report: tools.verify.go-junit-report
# 	@echo "===========> Run unit test"
# 	@$(GO) test -race -cover -coverprofile=$(OUTPUT_DIR)/coverage.out \
# 		-timeout=10m -shuffle=on -short -v `go list ./pkg/hook/ |\
# 		egrep -v $(subst $(SPACE),'|',$(sort $(EXCLUDE_TESTS)))` 2>&1 | \
# 		tee >(go-junit-report --set-exit-code >$(OUTPUT_DIR)/report.xml)
# 	@sed -i '/mock_.*.go/d' $(OUTPUT_DIR)/coverage.out # remove mock_.*.go files from test coverage
# 	@$(GO) tool cover -html=$(OUTPUT_DIR)/coverage.out -o $(OUTPUT_DIR)/coverage.html

## go.test.cover: Run unit test with coverage
.PHONY: go.test.cover
go.test.cover: go.test.junit-report
	@touch $(OUTPUT_DIR)/coverage.out
	@$(GO) tool cover -func=$(OUTPUT_DIR)/coverage.out | \
		awk -v target=$(COVERAGE) -f $(ROOT_DIR)/scripts/coverage.awk

## imports: task to automatically handle import packages in Go files using goimports tool
.PHONY: go.imports
go.imports: tools.verify.goimports
	@$(TOOLS_DIR)/goimports -l -w $(SRC)

## lint: Run the golangci-lint
.PHONY: go.lint
go.lint: tools.verify.golangci-lint
	@echo "===========> Run golangci to lint source codes"
	@$(TOOLS_DIR)/golangci-lint run -c $(ROOT_DIR)/.golangci.yml $(ROOT_DIR)/...

## go.clean: Clean all builds
.PHONY: go.clean
go.clean:
	@echo "===========> Cleaning all builds OUTPUT_DIR($(OUTPUT_DIR)) AND BIN_DIR($(BIN_DIR))"
	@-rm -vrf $(OUTPUT_DIR) $(BIN_DIR)
	@echo "===========> End clean..."

## go.rmi: Clean the project and remove the docker images
go.rmi:
	@echo "===========> Cleaning all build output"
	@docker rmi -f $(SWAGGER_NAME)
	@docker rmi -f $(CORE_NAME)

## copyright.help: Show copyright help
.PHONY: go.help
go.help: scripts/make-rules/golang.mk
	$(call smallhelp)
