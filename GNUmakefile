TEST ?= ./...
TESTARGS ?=

PROVIDER_HOST ?= github.com
NAMESPACE ?= nautobot
NAME ?= nautobot
VERSION ?= 3.1.0
BINARY := terraform-provider-$(NAME)
OS_ARCH := $(shell go env GOOS)_$(shell go env GOARCH)
PLUGIN_DIR := $(HOME)/.terraform.d/plugins/$(PROVIDER_HOST)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)
LDFLAGS := -X 'main.version=$(VERSION)'

TF_PLUGIN_CACHE_DIR ?= $(CURDIR)/.terraform.d/plugin-cache

DOCKER_COMPOSE ?= docker compose
COMPOSE_NAUTOBOT_FILE ?= test/nautobot/docker-compose.yml
COMPOSE_CI_FILE ?= test/docker-compose.yml
COMPOSE_LOCAL = $(DOCKER_COMPOSE) --project-directory "$(CURDIR)" -f "$(COMPOSE_NAUTOBOT_FILE)"
COMPOSE_CI = $(COMPOSE_LOCAL) -f "$(COMPOSE_CI_FILE)"

NAUTOBOT_TEST_URL ?= http://localhost:8080
NAUTOBOT_TEST_TOKEN ?= 0123456789abcdef0123456789abcdef01234567

NAUTOBOT_VER ?= 3.1.8
PYTHON_VER ?= 3.12
GO_VERSION ?= $(shell awk '/^toolchain go/ {v=$$2; sub(/^go/, "", v)} /^go / {g=$$2} END {print v ? v : g}' go.mod)

TF_TOOL ?= $(if $(shell command -v tofu 2>/dev/null),opentofu,terraform)
TF_VERSION ?= 1.11.1
TF_CLI := $(if $(filter opentofu,$(TF_TOOL)),tofu,terraform)
TF_CLI_PATH ?= $(shell command -v $(TF_CLI) 2>/dev/null)
TF_TARGET ?= $(if $(filter opentofu,$(TF_TOOL)),with-opentofu,with-terraform)
TF_ACC_TERRAFORM_PATH ?= $(TF_CLI_PATH)
TF_ACC_PROVIDER_NAMESPACE ?= hashicorp
TF_ACC_PROVIDER_HOST ?= $(if $(filter opentofu,$(TF_TOOL)),registry.opentofu.org,registry.terraform.io)

export GO_VERSION NAUTOBOT_VER PYTHON_VER TF_TOOL TF_VERSION TF_TARGET
export NAUTOBOT_TEST_URL NAUTOBOT_TEST_TOKEN
export TF_ACC_PROVIDER_HOST TF_ACC_PROVIDER_NAMESPACE

.DEFAULT_GOAL := install

.PHONY: build docs fmt fmt-check install local release release-check tag test testacc testacc-run \
	testacc-docker testacc-docker-down testacc-local testacc-local-up testacc-local-down

build:
	go build -ldflags "$(LDFLAGS)" -o "$(BINARY)"

fmt:
	gofmt -w $$(git ls-files '*.go')

fmt-check:
	@files="$$(gofmt -l $$(git ls-files '*.go'))"; \
	if [ -n "$$files" ]; then \
		echo "The following Go files are not formatted with gofmt:"; \
		echo "$$files"; \
		echo; \
		gofmt -d $$files; \
		echo "Run 'make fmt' and commit the result."; \
		exit 1; \
	fi

release:
	goreleaser build --snapshot --clean

install: build
	mkdir -p "$(PLUGIN_DIR)"
	cp "$(BINARY)" "$(PLUGIN_DIR)/$(BINARY)"

test:
	go test $(TESTARGS) -timeout 30s -parallel 4 $(TEST)

testacc-docker:
	@mkdir -p "$(TF_PLUGIN_CACHE_DIR)"
	@echo "Running acceptance tests via Docker Compose"
	@echo "  GO_VERSION=$(GO_VERSION)"
	@echo "  TF_TOOL=$(TF_TOOL) TF_VERSION=$(TF_VERSION) TF_TARGET=$(TF_TARGET)"
	@echo "  NAUTOBOT_VER=$(NAUTOBOT_VER) PYTHON_VER=$(PYTHON_VER)"
	@$(COMPOSE_CI) up --build --abort-on-container-exit --exit-code-from testrunner

testacc-docker-down:
	@$(COMPOSE_CI) down -v --remove-orphans

testacc-local-up:
	@echo "Starting reusable local Nautobot at $(NAUTOBOT_TEST_URL)"
	@$(COMPOSE_LOCAL) up -d --wait
	@echo "Local Nautobot is ready. Run 'make testacc-local-down' to remove it."

testacc-local-down:
	@$(COMPOSE_LOCAL) down -v --remove-orphans

testacc-run:
	@if [ -z "$(TF_ACC_TERRAFORM_PATH)" ]; then \
		echo "Error: '$(TF_CLI)' was not found. Set TF_ACC_TERRAFORM_PATH explicitly."; \
		exit 1; \
	fi
	@mkdir -p "$(TF_PLUGIN_CACHE_DIR)"
	@echo "Using $(TF_TOOL) ($(TF_ACC_TERRAFORM_PATH))"
	@TF_ACC=1 \
	  TF_ACC_TERRAFORM_PATH="$(TF_ACC_TERRAFORM_PATH)" \
	  TF_PLUGIN_CACHE_DIR="$(TF_PLUGIN_CACHE_DIR)" \
	  TF_ACC_PROVIDER_NAMESPACE="$(TF_ACC_PROVIDER_NAMESPACE)" \
	  TF_ACC_PROVIDER_HOST="$(TF_ACC_PROVIDER_HOST)" \
	  go test -p 1 $(TESTARGS) -timeout 120m -v $(TEST)

testacc-local: testacc-local-up
	@$(MAKE) testacc-run

testacc: testacc-local

local: install
	sed -i "s|version =.*|version = \"$(VERSION)\"|" local/provider.tf
	@if [ -z "$(TF_CLI_PATH)" ]; then \
		echo "Error: '$(TF_CLI)' was not found."; \
		exit 1; \
	fi
	cd local && "$(TF_CLI_PATH)" init -upgrade && "$(TF_CLI_PATH)" apply -auto-approve

docs:
	sed -i "s|version =.*|version = \"$(VERSION)\"|" README.md examples/provider/provider.tf local/provider.tf
	go generate ./...

release-check:
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Error: release checks require a clean working tree."; \
		git status --short; \
		exit 1; \
	fi
	@$(MAKE) fmt-check
	@$(MAKE) docs
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Error: generated documentation is not up to date."; \
		echo "Review and commit the following changes before tagging:"; \
		git status --short; \
		exit 1; \
	fi

tag: release-check
	@if git rev-parse "v$(VERSION)" >/dev/null 2>&1; then \
		echo "Error: tag v$(VERSION) already exists."; \
		exit 1; \
	fi
	git tag -a -m "Release v$(VERSION)" "v$(VERSION)"
	@echo "Created v$(VERSION). Push it explicitly with: git push origin v$(VERSION)"
