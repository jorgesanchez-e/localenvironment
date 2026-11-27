ROOT_DIR ?= $(shell git rev-parse --show-toplevel 2>/dev/null)
ifeq ($(ROOT_DIR),)
ROOT_DIR := $(abspath $(CURDIR)/../..)
endif

SUB_DIR := api/grpc
NOTIFIER_PROTO_DIR := proto/notifier
NOTIFIER_BUILD_DIR := $(ROOT_DIR)/pkg/grpc/notifier

all: notifier-v1

notifier-v1:
	buf generate --path $(NOTIFIER_PROTO_DIR) && \
	cd $(NOTIFIER_BUILD_DIR)/v1 && \
	go mod tidy

lint:
	buf lint	
	
clean:
	find $(NOTIFIER_BUILD_DIR)/v1 -type f -iname "*.go"  -exec rm -rf {} \;


.PHONY: all notifier-v1 lint 