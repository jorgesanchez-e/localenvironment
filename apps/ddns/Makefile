# Compatible with GNU Make and BSD Make (macOS, FreeBSD).
APPNAME ?= ddns
APPVERSION ?= 0.1
ROOT_DIR ?= $(shell git rev-parse --show-toplevel 2>/dev/null)
ifeq ($(ROOT_DIR),)
ROOT_DIR := $(.CURDIR)/../..
endif
GOOS ?=
GOARCH ?=

# used by `test` target
export REPORTS_DIR=./reports
# used by lint target
export GOLANGCILINT_VERSION=v2.12.2
export DOCKER_TAG=1.0

run: build
	ROOT_DIR=$(ROOT_DIR) GOOS=$(GOOS) GOARCH=$(GOARCH) APPVERSION=$(APPVERSION) APPNAME=$(APPNAME) ./scripts/run

build:
	ROOT_DIR=$(ROOT_DIR) GOOS=$(GOOS) GOARCH=$(GOARCH) APPVERSION=$(APPVERSION) APPNAME=$(APPNAME) ./scripts/build

test:
	ROOT_DIR=$(ROOT_DIR) APPNAME=$(APPNAME) ./scripts/unit-test

test-report:
	ROOT_DIR=$(ROOT_DIR) APPNAME=$(APPNAME) ./scripts/show-tests

lint:
	GOLANGCILINT_VERSION=$(GOLANGCILINT_VERSION) ROOT_DIR=$(ROOT_DIR) APPNAME=$(APPNAME) ./scripts/lint

clean:
	ROOT_DIR=$(ROOT_DIR) APPNAME=$(APPNAME) ./scripts/clean

.PHONY: build run test test-report lint clean