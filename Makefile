include help.mk

# get root dir
ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))

.DEFAULT_GOAL := start-docker

.PHONY: update
update: ## pulls git repo
	@git -C ${ROOT_DIR} pull
	go mod tidy

.PHONY: test
test: ## run golang test (including integration tests)
	go test ${ROOT_DIR}/...

.PHONY: build
build: ## build calender-manager binary
	go build ${ROOT_DIR}/...

.PHONY: lint
lint: ## run golangci-lint
	golangci-lint run ${ROOT_DIR}... -E dupl -E gocyclo -E gosec -E misspell -E sqlclosecheck

.PHONY: install-hooks
install-hooks: ## install git hooks
	@echo Installing git hooks...
	@go run -C .githooks install.go

# --- Local K3D development targets ---
IMAGE_NAME := calender-manager
IMAGE_VERSION := latest

.PHONY: start-cluster
start-cluster: ## starts k3d cluster and registry
	@k3d cluster create --config ${ROOT_DIR}k3d/clusterconfig.yaml

.PHONY: stop-k3d
stop-k3d: ## stop K3d
	@k3d cluster delete --config ${ROOT_DIR}k3d/clusterconfig.yaml

.PHONY: restart-k3d
restart-k3d: stop-k3d start-k3d ## restarts K3d

.PHONY: push-k3d
push-k3d: ## build and push server + scheduler images to local k3d registry
	@docker build --build-arg CMD=server ${ROOT_DIR} -t ${IMAGE_NAME}
	@docker tag ${IMAGE_NAME} localhost:5000/${IMAGE_NAME}:${IMAGE_VERSION}
	@docker push localhost:5000/${IMAGE_NAME}:${IMAGE_VERSION}

.PHONY: start-k3d
start-k3d: start-cluster push-k3d ## start k3d cluster and deploy operator + calender-manager CR
	@helm upgrade --install ${IMAGE_NAME} ${ROOT_DIR}charts/${IMAGE_NAME} \
		-f ${ROOT_DIR}k3d/values.k3d.yaml \
		--set server.image.repository=registry.localhost:5000/${IMAGE_NAME} \
		--set server.image.tag=${IMAGE_VERSION}

.PHONY: generate-helm-docs
generate-helm-docs: ## re-generates helm docs using docker
	@docker run --rm --volume "$(ROOT_DIR)charts:/helm-docs" jnorwood/helm-docs:latest