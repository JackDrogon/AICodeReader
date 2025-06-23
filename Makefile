# set makefile echo back
ifdef VERBOSE
	V :=
else
	V := @
endif

tag := $(shell git describe --abbrev=0 --always --dirty --tags)
sha := $(shell git rev-parse --short HEAD)
git_tag_sha := $(tag):$(sha)
LDFLAGS="-X 'github.com/JackDrogon/project/pkg/version.GitTagSha=$(git_tag_sha)'"
GOFLAGS=

# COVERAGE=ON to enable coverage
ifeq ($(COVERAGE),ON)
    GOFLAGS += -cover
endif
.PHONY: default
## default: Build aicodereader
default: aicodereader

.PHONY: build
## build : Build binaries
build: aicodereader

.PHONY: bin
bin:
	$(V)mkdir -p bin

.PHONY: lint
## lint : Lint codespace
lint:
	$(V)golangci-lint run --config .golangci.yml

.PHONY: fmt
## fmt : Format all code
fmt:
	$(V)go fmt ./...

.PHONY: ci
## ci : Run all CI checks locally
ci: check-fmt lint test build
	@echo "All CI checks passed!"

.PHONY: test
## test : Run test
test:
	$(V)go test -v -race -covermode=atomic -coverprofile=coverage.out ./...

.PHONY: test-coverage
## test-coverage : Run test with coverage report
test-coverage: test
	$(V)go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: check-fmt
## check-fmt : Check code formatting
check-fmt:
	$(V)if [ "$$(gofmt -s -l . | wc -l)" -gt 0 ]; then \
		echo "The following files are not formatted correctly:"; \
		gofmt -s -l .; \
		echo "Please run 'make fmt' to format your code."; \
		exit 1; \
	else \
		echo "All code is formatted correctly."; \
	fi

.PHONY: cloc
## cloc : Count lines of code
cloc:
	$(V)tokei -C .

.PHONY: todos
## todos : Print all todos
todos:
	$(V)grep -rnw . -e "TODO" | grep -v '^./pkg/rpc/thrift' | grep -v '^./.git'

.PHONY: check-env
## check-env : Check development environment
check-env:
	$(V)bash scripts/check-dev-env.sh

.PHONY: help
## help : Print help message
help: Makefile
	@sed -n 's/^##//p' $< | awk 'BEGIN {FS = ":"} {printf "\033[36m%-23s\033[0m %s\n", $$1, $$2}'
# --------------- ------------------ ---------------
# --------------- User Defined Tasks ---------------
.PHONY: aicodereader
## aicodereader : Build project
aicodereader: bin
	$(V)go build -ldflags $(LDFLAGS) -o bin/aicodereader ./cmd/aicodereader