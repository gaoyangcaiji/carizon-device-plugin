.PHONY: all push build
.DEFAULT_GOAL := all

##### Global variables #####

DOCKER   ?= docker
REGISTRY ?= hub.hobot.cc/dlp
VERSION  ?= 0.0.1
PKG ?= carizon-device-plugin

##### Public rules #####

all: build push

push:
	$(DOCKER) push "$(REGISTRY)/$(PKG):$(VERSION)"

build:
	./build.sh
	$(DOCKER) build \
		--tag $(REGISTRY)/$(PKG):$(VERSION) \
		--file Dockerfile .

clean:
	rm -f $(PKG)
	$(DOCKER) rmi "$(REGISTRY)/$(PKG):$(VERSION)"
