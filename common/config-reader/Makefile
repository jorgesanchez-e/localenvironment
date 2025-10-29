APPNAME ?= common/config-reader

ROOT_DIR = `git rev-parse --show-toplevel`

test:
	APPNAME=$(APPNAME) $(ROOT_DIR)/scripts/unit-test

test-report:
	APPNAME=$(APPNAME) $(ROOT_DIR)/scripts/show-tests

lint:
	APPNAME=$(APPNAME) $(ROOT_DIR)/scripts/lint

clean:
	APPNAME=$(APPNAME) $(ROOT_DIR)/scripts/clean

.PHONY: test test-report clean