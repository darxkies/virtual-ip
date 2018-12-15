VERSION = $(shell git describe --tags)
PACKAGE = github.com/darxkies/virtual-ip
BUILD_IMAGE = darxkies/virtual-ip-build
DEPLOY_IMAGE = darxkies/virtual-ip:$(VERSION)

compile:
	docker build -t $(BUILD_IMAGE) -f docker/build.dockerfile .
	docker run --rm -v $$(pwd):/go/src/$(PACKAGE) $(BUILD_IMAGE)

deploy: compile
	docker build -t $(DEPLOY_IMAGE) -f docker/deploy.dockerfile .
	docker push $(DEPLOY_IMAGE)

build-binaries:
	CGO_ENABLED=0 go build -ldflags "-X ${PACKAGE}/version.Version=${VERSION} -s -w" -o virtual-ip ${PACKAGE}/cmd/virtual-ip

watch-and-compile:
	go get github.com/cespare/reflex
	reflex -r '\.go$$' -R '^vendor' -R '^utils/a_utils-packr\.go$$' make build-binaries

clean:
	sudo rm -Rf bin vendor

.PHONY: build clean
