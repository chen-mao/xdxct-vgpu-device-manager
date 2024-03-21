MODULE := github.com/chen-mao/xdxct-vgpu-device-manager

DOCKER ?= docker

include $(CURDIR)/versions.mk

BUILDIMAGE ?= vgpu-device-manager-build

CMDS := $(patsubst ./cmd/%/,%,$(sort $(dir $(wildcard ./cmd/*/))))
CMD_TARGETS := $(patsubst %,cmd-%, $(CMDS))

CHECK_TARGETS := assert-fmt vet lint misspell ineffassign
MAKE_TARGETS := build fmt cmds $(CHECK_TARGETS)
TARGETS := $(MAKE_TARGETS)
DOCKER_TARGET := $(patsubst %, docker-%, $(TARGETS)) 


.PHONY: $(TARGETS) $(DOCKER_TARGETS)

GOOS := linux

cmds: $(CMD_TARGETS)
$(CMD_TARGETS): cmd-%:
	GOOS=$(GOOS) go build -ldflags "-s -w" $(COMMAND_BUILD_OPTIONS) $(MODULE)/cmd/$(*)

build:
	GOOS=$(GOOS) go build $(MODULE)/...

# Apply go fmt to the codebase
fmt:
	go list -f '{{.Dir}}' $(MODULE)/... \
		| xargs gofmt -s -l -w

assert-fmt:
	go list -f '{{.Dir}}' $(MODULE)/... \
		| xargs gofmt -s -l | ( grep -v /vendor/ || true ) > fmt.out
	@if [ -s fmt.out ]; then \
		echo "\nERROR: The following files are not formatted:\n"; \
		cat fmt.out; \
		rm fmt.out; \
		exit 1; \
	else \
		rm fmt.out; \
	fi


ineffassign:
	ineffassign $(MODULE)/...

lint:
# We use `go list -f '{{.Dir}}' $(MODULE)/...` to skip the `vendor` folder.
	go list -f '{{.Dir}}' $(MODULE)/... | xargs golint -set_exit_status

misspell:
	misspell $(MODULE)/...

vet:
	go vet $(MODULE)/...

.PHONY: .build-image .pull-build-image .push-build-image
.build-image: docker/Dockerfile.devel
	if [ x"$(SKIP_IMAGE_BUILD)" = x"" ]; then \
		$(DOCKER) build \
			--progress=plain \
			--build-arg GOLANG_VERSION="$(GOLANG_VERSION)" \
			--tag $(BUILDIMAGE) \
			-f $(^) \
			docker; \
	fi

.pull-build-image:
	$(DOCKER) pull $(BUILDIMAGE)

.push-build-image:
	$(DOCKER) push $(BUILDIMAGE)


$(DOCKER_TARGET): docker-%: .build-image
	@echo "RUN make $(*) in container $(BUILDIMAGE)"
	$(DOCKER) run \
	  --rm \
	  -e GOCACHE=/tmp/.cache \
	  -v $(PWD):$(PWD) \
	  -w $(PWD) \
	  --user $$(id -u):$$(id -g) \
	  $(BUILDIMAGE) \
	     make $(*)


