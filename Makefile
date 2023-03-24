# Copyright (c) 2023, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

NAME:=kontainer-engine-driver-ociocne

DRIVER_NAME:=kontainer-engine-driver-ociocne

# local build, use user and timestamp it
BINARY_NAME ?= ${NAME}
VERSION:=$(shell  date +%Y%m%d%H%M%S)
DIST_DIR:=dist
GO ?= go

#
# Go build related tasks
#
.PHONY: go-install
go-install:
	GO111MODULE=on $(GO) install .

.PHONY: go-run
go-run: go-install
	GO111MODULE=on $(GO) run .

.PHONY: go-fmt
go-fmt:
	gofmt -s -e -d $(shell find . -name "*.go" | grep -v /vendor/)

.PHONY: go-vet
go-vet:
	echo $(GO) vet $(shell $(GO) list ./... | grep -v /vendor/)

.PHONY: build
build:
	rm -rf ${DIST_DIR}
	mkdir -p ${DIST_DIR}
	GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o ${DIST_DIR}/${BINARY_NAME}-linux .

#
# Tests-related tasks
#
.PHONY: unit-test
unit-test: go-install
	go test -v ./...

.PHONY: ci
ci: unit-test build
