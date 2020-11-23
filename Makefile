VERSION = $(shell git describe --tags)
PACKAGE = github.com/darxkies/virtual-ip
BUILD_IMAGE = darxkies/virtual-ip-build
DEPLOY_IMAGE = darxkies/virtual-ip:$(VERSION)
DEPLOY_IMAGE_LATEST = darxkies/virtual-ip:latest
NAME?=virtual-ip-node00
IP=$(shell multipass info $(NAME) 2> /dev/null | grep IPv4 | sed -e "s/IPv4://" | xargs)
SSH_KEY=$(shell cat ~/.ssh/id_rsa.pub)
CLOUD_INIT="disable_root: 0\\nssh_authorized_keys:\\n  - $(SSH_KEY)" 
NODES=virtual-ip-node00 virtual-ip-node01 virtual-ip-node02
VIRTUAL_IP_SUFFIX=50
VIRTUAL_IP=$$($(MAKE) -s cluster-node-ip NAME=$(NAME) | sed -r 's/\.[0-9]+$$//').$(VIRTUAL_IP_SUFFIX)
SSH=ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null root@$(IP)
SCP=scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null 
INTERFACE=ens4 

compile:
	docker build -t $(BUILD_IMAGE) -f docker/build.dockerfile .
	docker run --rm -v $$(pwd):/go/src/$(PACKAGE) $(BUILD_IMAGE)

build-image: compile
	docker build -t $(DEPLOY_IMAGE) -f docker/deploy.dockerfile .
	docker tag $(DEPLOY_IMAGE) $(DEPLOY_IMAGE_LATEST)

deploy: build-image
	docker push $(DEPLOY_IMAGE)
	docker push $(DEPLOY_IMAGE_LATEST)

build-binaries:
	CGO_ENABLED=0 go build -ldflags "-X ${PACKAGE}/version.Version=${VERSION} -s -w" -o virtual-ip ${PACKAGE}/cmd/virtual-ip

install-reflex: 
	go get github.com/cespare/reflex

watch-and-compile: install-reflex
	reflex -r '\.go$$' -R '^vendor' -R '^utils/a_utils-packr\.go$$' $(MAKE) build-binaries

watch-compile-and-deploy: install-reflex
	reflex -r '\.go$$' -R '^vendor' -R '^utils/a_utils-packr\.go$$' $(MAKE) build-binaries cluster-virtual-ip-update

cluster-node-create:
	echo $(CLOUD_INIT) | multipass launch -n $(NAME) -c 2 -m 2G -d 20G --cloud-init - 20.04 

cluster-node-destroy:
	-multipass delete $(NAME) -p

cluster-node-start:
	-multipass start $(NAME) 

cluster-node-stop:
	-multipass stop $(NAME) 

cluster-node-inspect:
	$(SSH) ip neigh show
	$(SSH) ip a show dev $(INTERFACE)

cluster-node-virtual-ip-update:
	-$(SSH) rm /virtual-ip.deleted; mv /virtual-ip /virtual-ip.deleted
	-$(SCP) virtual-ip root@$(IP):/

cluster-node-virtual-ip-kill:
	-$(SSH) killall virtual-ip

cluster-node-ip:
	@echo $(IP)

cluster-node-peer:
	@echo "$(NAME)=$(IP):10000"

cluster-node-ping:
	IP=$(VIRTUAL_IP); $(SSH) ping -q -c 1 $$IP

cluster-node-virtual-ip-run:
	while [ true ]; do $(SSH) killall virtual-ip; $(SSH) /virtual-ip -id $(NAME) -bind $(IP):10000 -interface $(INTERFACE) -peers $$($(MAKE) -s cluster-peers) -virtual-ip $(VIRTUAL_IP); done

cluster-peers:
	@echo $(foreach node,$(NODES),$$($(MAKE) -s cluster-node-peer NAME=$(node))) | tr " " ","

cluster-create:
	for i in $(NODES); do $(MAKE) cluster-node-create NAME=$$i; done

cluster-destroy:
	for i in $(NODES); do $(MAKE) cluster-node-destroy NAME=$$i; done

cluster-start:
	for i in $(NODES); do $(MAKE) cluster-node-start NAME=$$i; done

cluster-stop:
	for i in $(NODES); do $(MAKE) cluster-node-stop NAME=$$i; done

cluster-virtual-ip-update:
	for i in $(NODES); do $(MAKE) cluster-node-virtual-ip-update NAME=$$i; done
	for i in $(NODES); do $(MAKE) cluster-node-virtual-ip-kill NAME=$$i; done

cluster-virtual-ip-ping:
	ping $(VIRTUAL_IP)

cluster-ping:
	for i in $(NODES); do $(MAKE) cluster-node-ping NAME=$$i; done

cluster-inspect:
	for i in $(NODES); do echo $$i; $(MAKE) -s cluster-node-inspect NAME=$$i; done

cluster-kill-random:
	while [ true ]; do $(MAKE) -s cluster-ping cluster-node-virtual-ip-kill NAME=$$(shuf -n1 -e $(NODES)); sleep $$(shuf -n1 -e 3 4 5); echo; done

clean:
	sudo rm -Rf bin vendor

.PHONY: build clean
