-include .env

BINS        := bin/newsletter sbin/newsletterctl
REMOTE      ?= club1.fr
REMOTE_PATH ?= /var/tmp/nlgo

all: $(BINS)

$(BINS):
	CGO_ENABLED=0 go build -o $@ ./cmd/$(@F)

deploy: $(BINS)
	rsync --checksum --archive --verbose $(BINS) $(REMOTE):$(REMOTE_PATH)

.PHONY: all $(BINS) deploy
