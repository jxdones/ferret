TARGET := ferret
VERSION := 0.2.1
GO := go
GOFMT := gofmt
LINTER := golangci-lint
PREFIX ?= /usr/local
GOOS := $(shell go env GOOS)

.PHONY: build test fmt lint clean release install install-prefix

LDFLAGS := -ldflags "-s -w -X main.version=v$(VERSION)"

build:
	$(GO) build $(LDFLAGS) -o bin/$(TARGET) .

test:
	$(GO) test ./...

fmt:
	$(GOFMT) -s -w .

lint:
	$(LINTER) run ./...

clean:
	rm -rf bin dist

release: clean
	mkdir -p dist
	GOARCH=amd64 $(GO) build $(LDFLAGS) -o dist/$(TARGET)-$(GOOS)-amd64 .
	GOARCH=arm64 $(GO) build $(LDFLAGS) -o dist/$(TARGET)-$(GOOS)-arm64 .

# Install to $GOBIN. Ensure $GOBIN is in your PATH.
install:
	$(GO) install $(LDFLAGS) .

install-prefix: build
	install -d $(DESTDIR)$(PREFIX)/bin
	install -m 755 bin/$(TARGET) $(DESTDIR)$(PREFIX)/bin/$(TARGET)
