-include .env

VERSION     ?= $(shell git describe --tags --always)
BINS        := bin/newsletter sbin/newsletterctl
REMOTE      ?= club1.fr
REMOTE_PATH ?= /var/tmp/nlgo

all: $(BINS)

$(BINS):
	CGO_ENABLED=0 go build -o $@ -ldflags '-X main.version=$(VERSION)' ./cmd/$(@F)

check: lint test

lint:
	! gofmt -s -d . | grep ''
	go vet ./...

test:
	go test -cover ./...

clean:
	rm -rf bin sbin

deploy: $(BINS)
	rsync --checksum --recursive --verbose bin sbin $(REMOTE):$(REMOTE_PATH)

.PHONY: all $(BINS) check lint test clean deploy
