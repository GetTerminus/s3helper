# Package information
PACKAGE   = s3helper
VERSION   = $(shell $(SEMANTICS) --dry-run --output-tag)
PLATFORMS = linux darwin
os        = $(word 1, $@)

# GitHub
GITHUB_ORG = GetTerminus

# Setup Go
GOPATH = $(shell go env GOPATH)
BIN    = $(GOPATH)/bin
BASE   = $(GOPATH)/src/github.com/$(GITHUB_ORG)/$(PACKAGE)
PKGS   = $(shell cd $(BASE) && env GOPATH=$(GOPATH) $(GO) list ./... | grep -v /vendor)

export GOPATH

# Go commands
GO ?= go

M = $(shell printf "\033[34;1m▶\033[0m")

# Print commands if VERBOSE env var is set
# VERBOSE ?= 0
Q := $(if $(VERBOSE),,@)

# Create releases for osx and linux
.PHONY: build
build: linux darwin

.PHONY: $(PLATFORMS)
$(PLATFORMS): $(BASE); $(info $(M) building executable for $(os)…) @
	$Q	cd $(BASE) && \
			env GOOS=$(os) GOARCH=amd64 $(GO) build \
				-o bin/$(PACKAGE)-$(os)-amd64

# If the repo is not in the proper dir in the GOPATH...
# create that dir and add a symlink the repo to it
$(BASE):
	$Q mkdir -p $(dir $@)
	$Q ln -sf $(CURDIR) $@

# Third party tools
$(BIN):
	$Q mkdir -p $@
$(BIN)/%: $(BIN)
	$Q $(GO) get $(REPOSITORY) || ret=$$?; exit $$ret

GOMETALINTER = $(BIN)/gometalinter
$(BIN)/gometalinter: REPOSITORY=github.com/alecthomas/gometalinter

# Useful commands

.PHONEY: lint
lint: $(GOMETALINTER) | $(BASE); $(info $(M) running gometalinter…)
	$Q $(GOMETALINTER) --install
	$Q cd $(BASE) && $(GOMETALINTER) ./... --vendor --deadline=300s

.PHONY: clean
clean: $(BASE); $(info $(M) cleaning…) @
	$Q cd $(BASE) && rm -rf bin
