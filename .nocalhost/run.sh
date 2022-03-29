#! /bin/sh

go run -mod=vendor core/main.go --config=/home/appops/config --roles=/home/appops/roles --regions=/home/appops/regions --environment=production
