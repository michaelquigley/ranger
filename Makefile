.DEFAULT_GOAL := build
GOBIN ?= $(shell go env GOPATH)/bin

ifeq ($(filter-out /,$(abspath $(GOBIN))),)
$(error GOBIN is '$(GOBIN)'; it must name a real directory)
endif

.PHONY: build test clean frontend frontend-test headless generate push

# build the embedded single-page UI into ui/dist.
frontend:
	npm --prefix ui install
	npm --prefix ui run build

frontend-test: frontend
	npm --prefix ui run test

# depends on frontend so go:embed always has content.
build: frontend
	go install ./...

# install a headless binary without requiring the embedded UI.
headless:
	go install -tags no_ui ./...

# the full gate includes the frontend because the shipped Go build embeds it.
test: frontend-test
	go test ./... -count=1
	go vet ./...

clean:
	go clean ./...
	rm -f "$(GOBIN)"/*
	rm -f ranger
	rm -rf ui/dist ui/node_modules

# generate both sides of the committed API contract: the ogen server (Go) and
# the TypeScript client types.
generate:
	go generate ./internal/api/
	npm --prefix ui install
	npm --prefix ui run gen:api

push: build
	push vendor "$(GOBIN)/ranger" ranger
