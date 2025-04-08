#!/bin/bash

pushd web
ng build --configuration production --aot
popd

go build -tags=prod -ldflags "-X main.Version=${1}" -o mkdocs-cms main.go