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
TEST_LINT_DIRS := config $(addprefix apps/,$(APPS))

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
	@for dir in $(TEST_LINT_DIRS); do \
		echo "==> Testing $$dir"; \
		$(MAKE) -C $$dir test ROOT_DIR=$(ROOT_DIR) || exit 1; \
	done

lint:
	@for dir in $(TEST_LINT_DIRS); do \
		echo "==> Linting $$dir"; \
		$(MAKE) -C $$dir lint ROOT_DIR=$(ROOT_DIR) || exit 1; \
	done

test-report:
	@for dir in $(TEST_LINT_DIRS); do \
		echo "==> Test report $$dir"; \
		$(MAKE) -C $$dir  test-report ROOT_DIR=$(ROOT_DIR) || exit 1; \
	done
