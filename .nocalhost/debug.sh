#! /bin/sh

dlv --headless --log --listen :9009 --api-version 2 --accept-multiclient --build-flags="-mod=vendor" debug core/main.go -- -config=/home/appops/config -roles=/home/appops/roles -regions=/home/appops/regions -environment=production
