REGISTRY              := your-registry
IMAGE_PREFIX          := $(REGISTRY)
REPO_ROOT             := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
IMAGE_TAG             := $(shell cat VERSION)
EFFECTIVE_VERSION     := $(IMAGE_TAG)-$(shell git rev-parse HEAD)
.PHONY: build
build:
	@go build -o bin/gardener-extension-csi-driver-synology ./cmd/gardener-extension-csi-driver-synology
.PHONY: docker-build
docker-build:
	@docker build -t $(IMAGE_PREFIX)/gardener-extension-csi-driver-synology:$(IMAGE_TAG) .
.PHONY: docker-push
docker-push:
	@docker push $(IMAGE_PREFIX)/gardener-extension-csi-driver-synology:$(IMAGE_TAG)
.PHONY: clean
clean:
	@rm -rf bin/
.PHONY: verify
verify:
	@go fmt ./...
	@go vet ./...
.PHONY: test
test:
	@go test ./...
.PHONY: tidy
tidy:
	@go mod tidy
