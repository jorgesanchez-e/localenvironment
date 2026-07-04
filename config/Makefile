# Compatible with GNU Make and BSD Make (macOS, FreeBSD).
APPNAME ?= config
ROOT_DIR ?= $(shell git rev-parse --show-toplevel 2>/dev/null)
ifeq ($(ROOT_DIR),)
ROOT_DIR := $(.CURDIR)/../..
endif
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# used by `test` target
export REPORTS_DIR=./reports
# used by lint target
export GOLANGCILINT_VERSION=v2.12.2

test:
	ROOT_DIR=$(ROOT_DIR) APPNAME=$(APPNAME) ./scripts/unit-test

test-report:
	ROOT_DIR=$(ROOT_DIR) APPNAME=$(APPNAME) ./scripts/show-tests

lint:
	GOLANGCILINT_VERSION=$(GOLANGCILINT_VERSION) ROOT_DIR=$(ROOT_DIR) APPNAME=$(APPNAME) ./scripts/lint

.PHONY: test test-report lint clean