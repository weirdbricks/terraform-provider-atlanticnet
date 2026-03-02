HOSTNAME=registry.terraform.io
NAMESPACE=weirdbricks
NAME=atlanticnet
VERSION=0.1.0
OS=$(shell go env GOOS)
ARCH=$(shell go env GOARCH)
INSTALL_DIR=~/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS)_$(ARCH)
BINARY=terraform-provider-$(NAME)

default: build

# ── Build & Install ───────────────────────────────────────────────────────────

.PHONY: build
build:
	go build -ldflags="-X main.version=$(VERSION)" -o $(BINARY) .

.PHONY: install
install: build
	mkdir -p $(INSTALL_DIR)
	cp $(BINARY) $(INSTALL_DIR)/$(BINARY)_v$(VERSION)
	@echo "Installed to $(INSTALL_DIR)"

# ── Testing ───────────────────────────────────────────────────────────────────

# Run all unit tests (no API key needed)
.PHONY: test
test:
	go test ./... -v -count=1 -timeout 60s

# Run acceptance tests against the real Atlantic.Net API
# Requires: TF_ACC=1, ATLANTICNET_ACCESS_KEY, ATLANTICNET_PRIVATE_KEY
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v -count=1 -timeout 30m $(TESTARGS)

# Run server acceptance tests (creates real billable VMs)
.PHONY: testacc-servers
testacc-servers:
	TF_ACC=1 ATLANTICNET_RUN_SERVER_TESTS=1 go test ./... -v -count=1 -timeout 60m $(TESTARGS)

# ── Code Quality ──────────────────────────────────────────────────────────────

.PHONY: vet
vet:
	go vet ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: lint
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not found; install from https://golangci-lint.run" && exit 1)
	golangci-lint run ./...

# ── Documentation ─────────────────────────────────────────────────────────────

.PHONY: docs
docs:
	@which tfplugindocs > /dev/null || go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
	tfplugindocs generate

# ── Release ───────────────────────────────────────────────────────────────────

.PHONY: release
release:
	@echo "Tag the release and push: git tag v$(VERSION) && git push origin v$(VERSION)"
	@echo "GoReleaser will build cross-platform binaries automatically via GitHub Actions."

.PHONY: clean
clean:
	rm -f $(BINARY)
