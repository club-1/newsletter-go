-include .env

BINS        := newsletter newsletterctl
REMOTE      ?= club1.fr
REMOTE_PATH ?= /var/tmp/nlgo

all: $(BINS)

$(BINS):
	CGO_ENABLED=0 go build ./cmd/$@

deploy: $(BINS)
	rsync --checksum --archive --verbose $(BINS) $(REMOTE):$(REMOTE_PATH)

.PHONY: all $(BINS) deploy
