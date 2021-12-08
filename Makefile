-include .env

TARGETS := i18n

#VERSION := $(shell git describe --tags)
VERSION := 0.1.1
BUILD := $(shell git rev-parse --short HEAD)
DATE := $(shell date +%Y-%m-%dT%TZ%z)
PROJECT_NAME := $(shell basename "$(PWD)")

# Go
GO_PROXY := https://goproxy.cn
GO_BIN := $(shell pwd)/bin
GO_FILES := $(wildcard *.go)
GO_MODULE := github.com/master-g/i18n

GOLANGCILINT := $(GO_BIN)/golangci-lint
GOLANGCILINT_VER := v1.42.1

# output
BIN := $(shell pwd)/bin

# Redirect error output
# STDERR := /tmp/.$(PROJECT_NAME)-stderr.txt

# Make is verbose in Linux. Make it silent.
# MAKEFLAGS += --silent


.PHONY: help
all: help
help: Makefile
	@echo
	@echo " Choose a command run in "$(PROJECT_NAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo


## mod: Reset go mod
.PHONY: mod
mod:
	@echo "  >  Reset go mod..."
	@rm go.mod go.sum
	@go mod init $(GO_MODULE)


## vendor: Module cleanup and vendor
.PHONY: vendor
vendor:
	@echo "  >  Module tidy and vendor..."
	@GOPROXY=$(GO_PROXY) go mod tidy
	@GOPROXY=$(GO_PROXY) go mod download


## lint: Lint go source files
$(GOLANGCILINT):
	@GOBIN=$(GO_BIN) wget -O - -q https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s $(GOLANGCILINT_VER)


.PHONY: lint
lint: $(GOLANGCILINT)
	@echo "  >  Linting..."
	@$(GO_BIN)/golangci-lint run ./...


## fmt: Formats go source files
.PHONY: fmt
fmt:
	@echo "  >  Formating..."
	@find . -type f -name '*.go' -not -path './vendor/*' -not -path './api/*' -not -path './pb/*' -not -path './.idea/*' -print0 | xargs -0 goimports -w


## build: Build all executables
.PHONY: build
build: $(TARGETS)


## release: Build executable files for macOS, Windows, Linux
.PHONY: release
release:
	@echo "  >  Releasing..."
	@GOBIN=$(GOBIN) \
	gox -ldflags \
	"-X $(GOMODULE)/pkg/version.Version=$(VERSION) \
	-X $(GOMODULE)/pkg/version.BuildDate=$(DATE) \
	-X $(GOMODULE)/pkg/version.CommitHash=$(BUILD)" \
	-osarch="darwin/amd64" \
	-osarch="windows/amd64" \
	-osarch="linux/amd64" \
	-output="release/{{.OS}}_{{.Arch}}/{{.Dir}}" ./cmd/i18n


.PHONY: dist
dist:
	@echo "  > Distributing..."
	@tar -zcvf "./release/$(TARGETS)_$(VERSION)_linux_amd64.tar.gz" -C ./release/linux_amd64/ .
	@zip -j "./release/$(TARGETS)_$(VERSION)_darwin_amd64.zip" "./release/darwin_amd64/$(TARGETS)"
	@zip -j "./release/$(TARGETS)_$(VERSION)_windows_amd64.zip" "./release/windows_amd64/$(TARGETS).exe"


## clean: Cleaning build cache
.PHONY: clean
clean:
	@echo "  >  Cleaning build cache..."
	@-for target in $(TARGETS); do rm -f $(BIN)/$$target; done;


$(TARGETS):
	@echo "  >  Building $@..."
	@-go build -o $(BIN)/$@ ./cmd/$@

