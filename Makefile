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

# Build all by default, even if it's not first
.DEFAULT_GOAL := help

.PHONY: all
all: tidy gen lint build

# ==============================================================================
# Build options

# TODO: It is recommended to refactor horizon cmd module with cobra and add version information. It should be placed in pkg directory
ROOT_PACKAGE=github.com/horizoncd/horizon
VERSION_PACKAGE=github.com/horizoncd/horizon/pkg/version

# ==============================================================================
# Includes
include scripts/make-rules/common.mk # make sure include common.mk at the first include line
include scripts/make-rules/golang.mk
include scripts/make-rules/image.mk
#include scripts/make-rules/deploy.mk
include scripts/make-rules/copyright.mk
include scripts/make-rules/gen.mk
#include scripts/make-rules/release.mk
#include scripts/make-rules/swagger.mk
#include scripts/make-rules/dependencies.mk
include scripts/make-rules/tools.mk

# ==============================================================================
# Usage

define USAGE_OPTIONS

Options:

  DEBUG            Whether or not to generate debug symbols. Default is 0.

  BINS             Binaries to build. Default is all binaries under cmd.
                   This option is available when using: make {build}(.multiarch)
                   Example: make build BINS="app"

  PLATFORMS        Platform to build for. Default is linux_arm64 and linux_amd64.
                   This option is available when using: make {build}.multiarch
                   Example: make build.multiarch PLATFORMS="linux_arm64 linux_amd64"

  V                Set to 1 enable verbose build. Default is 0.
endef
export USAGE_OPTIONS

# ==============================================================================
# Targets

## build: Build the project for horizon
.PHONY: build
build:
	@$(MAKE) go.build

## swagger: Build the swagger
.PHONY: swagger
swagger:
	@$(MAKE) image.swagger

## swagger-run: Run a swagger server locally
.PHONY: swagger-run
swagger-run: 
	@$(MAKE) go.swagger-run

## core: Build the core
.PHONY: core
core:
	@$(MAKE) image.core

## lint: Check syntax and styling of go sources.
.PHONY: lint
lint:
	@$(MAKE) go.lint

## gen: Generate all necessary files.
.PHONY: gen
gen:
	@$(MAKE) gen.run

## verify-copyright: Verify the license headers for all files.
.PHONY: verify-license
verify-license:
	@$(MAKE) copyright.verify

## add-license: Add copyright ensure source code files have license headers.
.PHONY: add-license
add-license:
	@$(MAKE) copyright.add

## tools: Install dependent tools.
.PHONY: tools
tools:
	@$(MAKE) tools.install

## test: Run unit test.
.PHONY: test
test:
	@$(MAKE) go.test

## cover: Run unit test and get test coverage.
.PHONY: cover 
cover:
	@$(MAKE) go.test.cover

## imports: task to automatically handle import packages in Go files using goimports tool
.PHONY: imports
imports:
	@$(MAKE) go.imports

## clean: Remove all files that are created by building. 
.PHONY: clean
clean:
	@$(MAKE) go.clean

## rmi: Clean the project and remove the docker images
.PHONY: rmi
rmi:
	@$(MAKE) go.rmi

## help: Show this help info.
.PHONY: help
help: Makefile
	$(call makehelp)

## all-help: Show all help details info.
.PHONY: all-help
all-help: go.help copyright.help tools.help image.help help
	$(call makeallhelp)
