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

IMAGES ?= 
SWAGGER_NAME := horizon-swagger
CORE_NAME := horizon-core

.PHONY: tools.install
tools.install: $(addprefix tools.install., $(TOOLS))

.PHONY: tools.install.%
tools.install.%:
	@echo "===========> Installing $*"
	@$(MAKE) install.$*

.PHONY: tools.verify.%
tools.verify.%:
	@if ! which $* &>/dev/null; then $(MAKE) tools.install.$*; fi


.PHONY: image.core
image.core:
ifeq ($(shell uname -m),arm64)
	@docker build -t $(CORE_NAME) -f build/core/Dockerfile . --platform linux/arm64
else
	@docker build -t $(CORE_NAME) -f build/core/Dockerfile .
endif

## swagger: Build the swagger
.PHONY: image.swagger
image.swagger:
ifeq ($(shell uname -m),arm64)
	@docker build -t $(SWAGGER_NAME) -f build/swagger/Dockerfile . --platform linux/arm64
else
	@docker build -t $(SWAGGER_NAME) -f build/swagger/Dockerfile .
endif

## image.help: Print help for image targets
.PHONY: image.help
image.help: scripts/make-rules/image.mk
	$(call smallhelp)