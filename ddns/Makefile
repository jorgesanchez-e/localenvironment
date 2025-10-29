APPNAME ?= ddns

ROOT_DIR=`git rev-parse --show-toplevel`
#export ROOT_DIR=${ROOT_DIR}

build: clean
	mkdir -p build
	GOOS=$(GOOS) GOARCH=$(GOARCH) APPNAME=$(APPNAME) ../scripts/build

run: build
	$(ROOT_DIR)/build/${APPNAME}

test:
	APPNAME=$(APPNAME) $(ROOT_DIR)/scripts/unit-test

test-report:
	APPNAME=$(APPNAME) $(ROOT_DIR)/scripts/show-tests

lint:
	APPNAME=$(APPNAME) $(ROOT_DIR)/scripts/lint

clean:
	APPNAME=$(APPNAME) $(ROOT_DIR)/scripts/clean

.PHONY: build run test test-report clean