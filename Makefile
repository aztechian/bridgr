PROJECT_NAME := bridgr
PKG := "$(PROJECT_NAME)"
CMD := cmd/bridgr/main.go
GO_FILES := $(shell find . -name '*.go' | grep -v _test.go)
OS := $(shell go env GOOS)
ARCH := $(shell go env GOARCH)

.SUFFIXES:
.PHONY: all coverage lint test junit race clean generate download locallint

ifeq ($(OS), linux)
PROJECT_NAME := $(PROJECT_NAME)-Linux-$(ARCH)
else ifeq ($(OS), windows)
PROJECT_NAME := $(PROJECT_NAME)-Windows-$(ARCH)
else ifeq ($(OS), darwin)
PROJECT_NAME := $(PROJECT_NAME)-MacOS-$(ARCH)
endif

all: $(PROJECT_NAME) $(PROJECT_NAME).sha256

ifeq ($(CI),)
lint: locallint
else
lint: cilint.txt
endif

coverage: coverage.out
ifeq ($(GITHUB_ACTIONS),)
	@go tool cover --html=$<
else
	@go tool cover --html=$< -o coverage.html
endif

ifdef GITHUB_REF_NAME
LDFLAGS := -ldflags="-X github.com/aztechian/bridgr/internal/bridgr.Version=$(GITHUB_REF_NAME)"
endif

locallint:
	@golangci-lint run

cilint.txt: $(GO_FILES)
	@golangci-lint run --out-format=line-number --new-from-rev=master --issues-exit-code=0 > $@

test:
	@go test -short ./...

race:
	@go test -v -count=1 -race ./...

coverage.out: $(GO_FILES)
	@go test -v -race -covermode=atomic -coverprofile=$@ ./... 2>&1 | tee tests.out

junit:
	@go install github.com/jstemmer/go-junit-report@latest

report.xml: coverage.out junit
	@go-junit-report -set-exit-code < tests.out > $@

run:
	@go run $(CMD) -c config/example.yml

clean:
	@rm -rf internal/bridgr/asset/templates.go coverage.out packages tests.out report.xml *.sha256 main cilint.txt $(PKG)-*
	@docker rm --force bridgr_yum bridgr_python bridgr_ruby &> /dev/null || true

generate: $(GO_FILES)
	@GOOS="" GOARCH="" go generate ./...

$(PROJECT_NAME): generate $(GO_FILES)
	@go build -tags dist -v -o $@ $(LDFLAGS) $(CMD)

%.sha256:
	@openssl dgst -sha256 -hex $* | cut -f2 -d' ' > $@

download: generate
	@go mod download
