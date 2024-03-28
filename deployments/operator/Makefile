DOCKER ?= docker

##### Global variables #####
include $(CURDIR)/versions.mk

DISTRIBUTIONS := dev

BUILD_TARGETS := $(patsubst %, build-%, $(DISTRIBUTIONS))
PUSH_TARGETS := $(patsubst %, push-%, $(DISTRIBUTIONS))

REGISTRY ?= hub.xdxct.com/kubevirt
IMAGE_NAME := $(REGISTRY)/xdxct-vgpu-device-manager
IMAGE_TAG := devel

IMAGE := $(IMAGE_NAME):$(IMAGE_TAG)

.PHONY: $(BUILD_TARGETS)

build-%: DOCKERFILE_SUFFIX = $(*)
build-%: DOCKERFILE = $(CURDIR)/deployments/operator/Dockerfile.$(DOCKERFILE_SUFFIX)
build-%: DOCKER_BUILD_PLATFORM_OPTIONS = --platform=linux/amd64

$(BUILD_TARGETS): build-%:
	$(DOCKER) build \
		--tag $(IMAGE) \
		--build-arg GOLANG_VERSION="$(GOLANG_VERSION)"
		--build-arg VERSION="$(VERSION)" \
		--file $(DOCKERFILE) \
		$(CURDIR)

$(PUSH_TARGETS):
	$(DOCKER) push $(IMAGE)
