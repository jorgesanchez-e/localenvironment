APPNAME ?= common/config-reader

ROOT_DIR = `git rev-parse --show-toplevel`

# used by lint target
export GOLANGCILINT_VERSION=v1.54.2

test:
	APPNAME=$(APPNAME) $(ROOT_DIR)/scripts/unit-test

test-report:
	APPNAME=$(APPNAME) $(ROOT_DIR)/scripts/show-tests

lint:
	APPNAME=$(APPNAME) $(ROOT_DIR)/scripts/lint

clean:
	APPNAME=$(APPNAME) $(ROOT_DIR)/scripts/clean

.PHONY: test test-report clean