.PHONY: build headless generate test clean push
.DEFAULT_GOAL := build
GOBIN ?= $(shell go env GOPATH)/bin

clean:
	go clean
	rm -f ${GOBIN}/vane vane
	rm -rf ui/dist ui/node_modules

build:
	npm --prefix ui install
	npm --prefix ui run build
	go install ./...

headless:
	go install -tags no_ui ./...

generate:
	go generate ./internal/api/
	npm --prefix ui run gen:api

test:
	go test ./... -count=1
	go vet ./...

push:
	push vendor ${GOBIN}/vane vane
