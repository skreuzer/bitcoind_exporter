CONTAINER_IMAGE_NAME?= bitcoind_exporter
CONTAINER_IMAGE_TAG?= $(subst /,-,$(shell git rev-parse --short HEAD))

container:
	@podman build -t "$(CONTAINER_IMAGE_NAME):$(CONTAINER_IMAGE_TAG)" -f ./Containerfile

test:
	echo $(DOCKER_IMAGE_TAG)
.PHONY: container
