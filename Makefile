SOURCE_VERSION = $(shell git rev-parse --short=6 HEAD)
BUILD_FLAGS = -v -ldflags "-X github.com/DECODEproject/iot-prototype.SourceVersion=$(SOURCE_VERSION)"
PACKAGES := $(shell go list ./... | grep -v /vendor/ )

GO_TEST = go test -covermode=atomic
GO_INTEGRATION = $(GO_TEST) -bench=. -v --tags=integration
GO_COVER = go tool cover
GO_BENCH = go test -bench=.
ARTEFACT_DIR = coverage

all: linux-arm linux-i386 linux-amd64 darwin-amd64 ## build executables for the various environments

.PHONY: all

get-build-deps: ## install build dependencies 
	go get github.com/chespinoza/goliscan

.PHONY: get-build-deps

check-vendor-licenses: ## check if licenses of project dependencies meet project requirements 
	@goliscan check --direct-only -strict
	@goliscan check --indirect-only -strict

.PHONY: check-vendor-licenses

test: ## run tests
	$(GO_TEST) $(PACKAGES)

.PHONY: test

test_integration: ## run integration tests (SLOW)
	mkdir -p $(ARTEFACT_DIR)
	echo 'mode: atomic' > $(ARTEFACT_DIR)/cover-integration.out
	touch $(ARTEFACT_DIR)/cover.tmp
	$(foreach package, $(PACKAGES), $(GO_INTEGRATION) -coverprofile=$(ARTEFACT_DIR)/cover.tmp $(package) && tail -n +2 $(ARTEFACT_DIR)/cover.tmp >> $(ARTEFACT_DIR)/cover-integration.out || exit;)
.PHONY: test_integration

clean: ## clean up
	rm -rf tmp/
	rm -rf $(ARTEFACT_DIR)

.PHONY: clean

bench: ## run benchmark tests
	$(GO_BENCH) $(PACKAGES)

.PHONY: bench

coverage: test_integration ## generate and display coverage report
	$(GO_COVER) -func=$(ARTEFACT_DIR)/cover-integration.out

.PHONY: test_integration

ui-build: ## build the ui project
	$(MAKE) -C ui build

.PHONY: ui-build

docker-build: clean ui-build linux-amd64 ## build docker images for all of the executables
	docker build -t prototype/metadata:latest -t prototype/metadata:$(SOURCE_VERSION) -f=./docker/Dockerfile.metadata .
	docker build -t prototype/node:latest -t prototype/node:$(SOURCE_VERSION) -f=./docker/Dockerfile.node .
	docker build -t prototype/storage:latest -t prototype/storage:$(SOURCE_VERSION) -f=./docker/Dockerfile.storage .

.PHONY: docker-build

docker-redis: ## run a local instance of docker for the storage service
	docker run -p 6379:6379 redis:3.0.7

.PHONY: docker-redis

client-metadata: ## build golang client for the metadata service - requires a local running metadata service
	java -jar ./tools/swagger-codegen-cli-2.2.1.jar generate -i http://localhost:8081/apidocs.json -l go -o ./client/metadata/ --type-mappings  object=interface{}

.PHONY: client-metadata

client-storage: ## build golang client for the storage service - requires a local running storage service
	java -jar ./tools/swagger-codegen-cli-2.2.1.jar generate -i http://localhost:8083/apidocs.json -l go -o ./client/storage/ --type-mappings  object=interface{}

.PHONY: client-storage

darwin-amd64: tmp/build/darwin-amd64/metadata tmp/build/darwin-amd64/storage tmp/build/darwin-amd64/node ## build for mac amd64

linux-i386: tmp/build/linux-i386/metadata tmp/build/linux-i386/storage tmp/build/linux-i386/node ## build for linux i386

linux-amd64: tmp/build/linux-amd64/metadata tmp/build/linux-amd64/storage tmp/build/linux-amd64/node ## build for linux amd64

linux-arm: tmp/build/linux-arm/metadata tmp/build/linux-arm/storage tmp/build/linux-arm/node ## build for linux arm (raspberry-pi)

.PHONY: darwin-amd64 linux-i386 linux-amd64 linux-arm

tmp/build/linux-i386/metadata:
	GOOS=linux GOARCH=386 go build $(BUILD_FLAGS) -o $(@) ./cmd/metadata

tmp/build/linux-i386/storage:
	GOOS=linux GOARCH=386 go build $(BUILD_FLAGS) -o $(@) ./cmd/storage

tmp/build/linux-i386/node:
	GOOS=linux GOARCH=386 go build $(BUILD_FLAGS) -o $(@) ./cmd/node

## linux-amd64
tmp/build/linux-amd64/metadata:
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $(@) ./cmd/metadata

tmp/build/linux-amd64/storage:
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $(@) ./cmd/storage

tmp/build/linux-amd64/node:
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $(@) ./cmd/node

## linux-arm
tmp/build/linux-arm/metadata:
	GOOS=linux GOARCH=arm go build $(BUILD_FLAGS) -o $(@) ./cmd/metadata

tmp/build/linux-arm/storage:
	GOOS=linux GOARCH=arm go build $(BUILD_FLAGS) -o $(@) ./cmd/storage

tmp/build/linux-arm/node:
	GOOS=linux GOARCH=arm go build $(BUILD_FLAGS) -o $(@) ./cmd/node

## darwin-amd64
tmp/build/darwin-amd64/metadata:
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o $(@) ./cmd/metadata

tmp/build/darwin-amd64/storage:
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o $(@) ./cmd/storage

tmp/build/darwin-amd64/node:
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o $(@) ./cmd/node


# 'help' parses the Makefile and displays the help text
help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: help
