SHELL   = /usr/bin/env bash
GIT_SHA = $(shell git rev-parse --short HEAD)
GIT_TAG = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)

VERSION = ${GIT_TAG}
ifeq ($(VERSION),)
	VERSION = ${GIT_SHA}-devel
endif

.PHONY: vendor
vendor:
	go mod vendor

.PHONY: build-magic
build-magic:
	CGO_ENABLED=0 \
		go build -ldflags="-X main.Version=$(VERSION)" \
			-o bin/docker-credential-magic \
			.../cmd/docker-credential-magic

.PHONY: test
test:
	scripts/test.sh

.PHONY: acceptance
acceptance:
	scripts/acceptance.sh

.PHONY: clean
clean:
	rm -rf .venv/ .cover/ .robot/ bin/ tmp/ vendor/
