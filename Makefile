SHELL=bash
MAIN=dp-table-renderer

BUILD_DIR ?= build
BUILD_ARCH=$(GOOS)-$(GOARCH)

BIN_DIR ?= $(BUILD_DIR)/$(BUILD_ARCH)

BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)

export GOOS=$(shell go env GOOS)
export GOARCH=$(shell go env GOARCH)

all: audit test build

audit:
	go list -json -m all | nancy sleuth --exclude-vulnerability-file ./.nancy-ignore

build:
	go build -tags 'production' -o $(BUILD_DIR)/dp-table-renderer -ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)"

debug:
	go build -tags 'debug' -race -o $(BUILD_DIR)/dp-table-renderer -ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)"
	HUMAN_LOG=1 DEBUG=1 $(BUILD_DIR)/dp-table-renderer

test:
	go test -cover $(shell go list ./... | grep -v /vendor/)

.PHONY: audit build debug test
