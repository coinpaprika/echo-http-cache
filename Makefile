SHELL := /bin/bash
export GO111MODULE=on
export GOPROXY=https://proxy.golang.org

# Pin golangci-lint version for reproducible local + CI runs.
# Bumped by Dependabot via .github/dependabot.yml (github-actions ecosystem
# tracks the version in .github/workflows/main.yml; mirror it here manually
# when that bumps).
GOLANGCI_VERSION := v2.11.4

.PHONY: check format help test tidy

format: ## Format go code with goimports
	@go install golang.org/x/tools/cmd/goimports@latest
	@goimports -l -w .

test: ## Run tests
	@go test -race ./...

tidy: ## Run go mod tidy
	@go mod tidy

check: ## Linting and static analysis
	@if test ! -x ./bin/golangci-lint || ! ./bin/golangci-lint --version | grep -q "$(patsubst v%,%,$(GOLANGCI_VERSION))"; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/$(GOLANGCI_VERSION)/install.sh \
		| sh -s -- -b ./bin $(GOLANGCI_VERSION); \
	fi

	@./bin/golangci-lint run -c .golangci.yml

	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@govulncheck ./...

help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
