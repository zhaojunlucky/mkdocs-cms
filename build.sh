#!/bin/bash

pushd web
ng build --configuration production --aot
popd

go build -tags=prod -o mkdocs-cms main.go