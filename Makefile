TARGETS := $(shell ls scripts | grep -v \\.sh)

.dapper:
	@echo Downloading dapper
	@https_proxy=10.130.0.16:7890 curl -sL https://github.com/Loongson-Cloud-Community/dapper/releases/download/v0.6.0/dapper-Linux-loong64 > .dapper.tmp
	@@chmod +x .dapper.tmp
	@mv .dapper.tmp .dapper

$(TARGETS): .dapper
	./.dapper $@

.PHONY: deps
deps:
	go mod vendor
	go mod tidy

release:
	./scripts/release.sh

.DEFAULT_GOAL := ci

.PHONY: $(TARGETS)

.PHONY: generate
generate: build/data 
	./scripts/download
	go generate

build/data:
	mkdir -p $@

.PHONY: binary-size-check
binary-size-check:
	scripts/binary_size_check.sh

.PHONY: image-scan
image-scan:
	scripts/image_scan.sh $(IMAGE)
