# Compatible with GNU Make and BSD Make (macOS, FreeBSD).
# Scripts use /bin/sh (POSIX). Requires: go, git. Lint also requires: curl.
ROOT_DIR ?= $(shell git rev-parse --show-toplevel 2>/dev/null)
ifeq ($(ROOT_DIR),)
ROOT_DIR := $(.CURDIR)
endif

GOOS ?=
GOARCH ?=
APPVERSION ?= 0.1

APPS := $(notdir $(patsubst %/Makefile,%,$(wildcard apps/*/Makefile)))

.PHONY: all build clean test lint test-report $(APPS)

all: build

build: $(APPS)

$(APPS):
	@echo "==> Building $@"
	@$(MAKE) -C apps/$@ build \
		ROOT_DIR=$(ROOT_DIR) \
		GOOS=$(GOOS) \
		GOARCH=$(GOARCH) \
		APPVERSION=$(APPVERSION)

clean:
	@for app in $(APPS); do \
		echo "==> Cleaning $$app"; \
		$(MAKE) -C apps/$$app clean ROOT_DIR=$(ROOT_DIR) || exit 1; \
	done

test:
	@for app in $(APPS); do \
		echo "==> Testing $$app"; \
		$(MAKE) -C apps/$$app test ROOT_DIR=$(ROOT_DIR) || exit 1; \
	done

lint:
	@for app in $(APPS); do \
		echo "==> Linting $$app"; \
		$(MAKE) -C apps/$$app lint ROOT_DIR=$(ROOT_DIR) || exit 1; \
	done

test-report:
	@for app in $(APPS); do \
		echo "==> Test report $$app"; \
		$(MAKE) -C apps/$$app test-report ROOT_DIR=$(ROOT_DIR) || exit 1; \
	done
