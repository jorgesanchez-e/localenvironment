APPNAME ?= ddns

ROOT_DOR=$(git rev-parse --show-toplevel)
export ROOT_DIR$(ROOT_DIR)

# used by lint target
export GOLANGCILINT_VERSION=v1.54.2

build: clean
	mkdir -p build
	GOOS=$(GOOS) GOARCH=$(GOARCH) APPNAME=$(APPNAME) ../scripts/build

run: build
	../build/${APPNAME}

test:
	APPNAME=$(APPNAME) ../scripts/unit-test

test-report:
	APPNAME=$(APPNAME) ../scripts/show-tests

lint:
	APPNAME=$(APPNAME) ../scripts/lint

clean:
	APPNAME=$(APPNAME) ../scripts/clean

.PHONY: build run test test-report clean