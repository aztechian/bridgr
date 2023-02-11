PROJECT_NAME := bridgr
PKG := "$(PROJECT_NAME)"
CMD := cmd/bridgr/main.go
GO_FILES := $(shell find . -name '*.go' | grep -v _test.go)
OS := $(shell go env GOOS)
ARCH := $(shell go env GOARCH)

.SUFFIXES:
.PHONY: all coverage lint test race x2unit xunit clean generate download locallint

ifeq ($(OS), linux)
PROJECT_NAME := $(PROJECT_NAME)-Linux-$(ARCH)
else ifeq ($(OS), windows)
PROJECT_NAME := $(PROJECT_NAME)-Windows-$(ARCH)
else ifeq ($(OS), darwin)
PROJECT_NAME := $(PROJECT_NAME)-MacOS-$(ARCH)
endif

all: $(PROJECT_NAME) $(PROJECT_NAME).sha256

ifeq ($(CI),)
coverage: html
lint: locallint
else
coverage: coverage.out
lint: cilint.txt
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
	@go test -v -race -covermode=atomic -coverprofile=$@ ./...

html: coverage.out
	@go tool cover --html=$<

x2unit:
	go get github.com/tebeka/go2xunit

tests.out:
	@go test -v -race ./... > $@

xunit: x2unit tests.out
	go2xunit -fail -input tests.out -output tests.xml
	@rm -f tests.out

run:
	@go run $(CMD) -c config/example.yml

clean:
	@rm -rf internal/bridgr/asset/templates.go coverage.out packages tests.xml tests.out coverage.out *.sha256 main cilint.txt $(PKG)-*
	@docker rm --force bridgr_yum bridgr_python bridgr_ruby &> /dev/null || true

generate: $(GO_FILES)
	@GOOS="" GOARCH="" go generate ./...

$(PROJECT_NAME): generate $(GO_FILES)
	@go build -tags dist -i -v -o $@ $(LDFLAGS) $(CMD)

%.sha256:
	@openssl dgst -sha256 -hex $* | cut -f2 -d' ' > $@

download: generate
	@go mod download
