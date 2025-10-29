# used by `test` target
export REPORTS_DIR=./reports
export ROOT_DIR=$(git rev-parse --show-toplevel)

COMPONENTS=ddns

all: build

build:
	for app in $(COMPONENTS); do \
		$(MAKE) -C $$app build; \
	done

clean:
	for app in $(COMPONENTS); do \
		$(MAKE) -C $$app clean; \
	done

test:
	for app in $(COMPONENTS); do \
		$(MAKE) -C $$app test; \
	done

lint:
	for app in $(COMPONENTS); do \
		$(MAKE) -C $$app lint; \
	done


.PHONY: all build clean test lint