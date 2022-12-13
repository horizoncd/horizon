#!/bin/bash

GOFLAGS="-mod=vendor" go test $(go list ./...) -coverprofile=coverage.data -timeout 30m >&2 || return 1

go tool cover -func=coverage.data >&2
COVERAGE=$(go tool cover -func=coverage.data | tail -n 1 | awk '{print $3}') && COVERAGE=${COVERAGE%%\%} && echo "$COVERAGE"
