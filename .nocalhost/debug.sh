#! /bin/sh

dlv --headless --log --listen :9009 --api-version 2 --accept-multiclient debug core/main.go -- -config=/home/appops/config -roles=/home/appops/roles -environment=production -scopes=/home/appops/scopes -buildjsonschema=/home/appops/buildjsonschema -builduischema=/home/appops/builduischema
