# Copyright 2023 The horizoncd Authors.
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
	$(error the variable ROOT_PACKAGE must be set prior to including golang.mk, ->/Makefile)
endif

GOPATH ?= $(shell go env GOPATH)
ifeq ($(origin GOBIN), undefined)
	GOBIN := $(GOPATH)/bin
endif

## go.build: Build binaries
.PHONY: go.build
go.build:
	@echo "COMMAND=$(COMMAND)"
	@echo "PLATFORM=$(PLATFORM)"
	@echo "OS=$(OS)"
	@echo "ARCH=$(ARCH)"
	@echo "GO=$(shell go version)"
	@echo "===========> Building binary $(BINS) $(VERSION) for $(PLATFORM)"
	@export CGO_ENABLED=0 && go build -o bin/app -ldflags '-s -w' ./core/main.go

## go.test: Run unit test
.PHONY: go.test
go.test: tools.verify.go-junit-report
	@echo "===========> Run unit test"


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
	@docker rmi -f $(JOB_NAME)
	@docker rmi -f $(CORE_NAME)

## copyright.help: Show copyright help
.PHONY: go.help
go.help: scripts/make-rules/golang.mk
	$(call smallhelp)