# Makefile for podhnologic

export PATH := /opt/homebrew/bin:/usr/local/bin:$(PATH)

BINARY_NAME=podhnologic
BUILD_DIR=build
GOCMD=go
GOTEST=$(GOCMD) test

.PHONY: all build build-all build-darwin-amd64 build-darwin-arm64 build-linux-amd64 build-linux-arm64 build-windows-amd64 build-windows-arm64 clean deps help install linked-test run test web-build

all: test build

build:
	@./scripts/build-linked.sh

build-all: build-darwin-amd64 build-darwin-arm64 build-linux-amd64 build-linux-arm64 build-windows-amd64

build-darwin-amd64:
	@./scripts/build-linked.sh darwin-amd64

build-darwin-arm64:
	@./scripts/build-linked.sh darwin-arm64

build-linux-amd64:
	@./scripts/build-linked.sh linux-amd64

build-linux-arm64:
	@./scripts/build-linked.sh linux-arm64

build-windows-amd64:
	@./scripts/build-linked.sh windows-amd64

build-windows-arm64:
	@./scripts/build-linked.sh windows-arm64

deps:
	@$(GOCMD) mod download
	@$(GOCMD) mod tidy
	@cd web && pnpm install --frozen-lockfile

test:
	@$(GOTEST) -v -count=1 ./...

linked-test: build
	@case "$$(uname -s):$$(uname -m)" in \
		Darwin:arm64|Darwin:aarch64) target="darwin-arm64"; cc="clang"; cflags="-arch arm64"; ldflags="-arch arm64" ;; \
		Darwin:x86_64) target="darwin-amd64"; cc="clang"; cflags="-arch x86_64"; ldflags="-arch x86_64" ;; \
		Linux:x86_64|Linux:amd64) target="linux-amd64"; cc="cc"; cflags=""; ldflags="" ;; \
		Linux:arm64|Linux:aarch64) target="linux-arm64"; cc="cc"; cflags=""; ldflags="" ;; \
		*) echo "unsupported linked-test host: $$(uname -s)/$$(uname -m)" >&2; exit 1 ;; \
	esac; \
	prefix="$(CURDIR)/scripts/ffmpeg/dist/$$target"; \
	export CC="$$cc"; \
	export CGO_ENABLED=1; \
	export PKG_CONFIG_PATH="$$prefix/lib/pkgconfig"; \
	export CGO_CFLAGS="$$cflags -I$$prefix/include"; \
	export CGO_LDFLAGS="$$ldflags -L$$prefix/lib -lpodhnologicffmpeg $$(pkg-config --static --libs libavfilter libavformat libavcodec libswresample libswscale libavutil)"; \
	$(GOTEST) -v -count=1 -tags 'linkedffmpeg_cgo linkedffmpeg_hidden' ./...

web-build:
	@cd web && pnpm build

clean:
	@$(GOCMD) clean
	@rm -rf $(BUILD_DIR)

install: build
	@cp "$$(ls -t $(BUILD_DIR)/$(BINARY_NAME)-* | head -1)" /usr/local/bin/$(BINARY_NAME)

run: build
	@"$$(ls -t $(BUILD_DIR)/$(BINARY_NAME)-* | head -1)"

help:
	@echo "Available targets:"
	@echo "  make build              Build linked binary for this host"
	@echo "  make build-all          Build configured native targets"
	@echo "  make test               Run Go tests"
	@echo "  make linked-test        Run Go tests with linked FFmpeg tags"
	@echo "  make web-build          Build the Astro web app"
	@echo "  make clean              Remove build artifacts"
	@echo "  make install            Install to /usr/local/bin"
