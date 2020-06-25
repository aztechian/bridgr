PROJECT_NAME := bridgr
PKG := "$(PROJECT_NAME)"
CMD := cmd/bridgr/main.go
VERSION := $(shell  git describe --always --dirty | sed 's/^v//')
GO_FILES := $(shell find . -name '*.go' | grep -v _test.go)

.SUFFIXES:
.PHONY: all coverage lint test race x2unit xunit clean generate download locallint cilint

ifeq ($(GOOS), linux)
PROJECT_NAME := $(PROJECT_NAME)-Linux
else ifeq ($(GOOS), windows)
PROJECT_NAME := $(PROJECT_NAME)-Windows
else ifeq ($(GOOS), darwin)
PROJECT_NAME := $(PROJECT_NAME)-MacOS
endif

all: $(PROJECT_NAME) $(PROJECT_NAME).sha256

ifeq ($(TRAVIS),)
coverage: html
lint: locallint
else
coverage: coverage.out
lint: cilint
endif

locallint:
	@golangci-lint run

cilint:
	@golangci-lint run --out-format=code-climate --new-from-rev=master --issues-exit-code=0

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
	@rm -rf internal/bridgr/asset/templates.go coverage.out packages tests.xml tests.out coverage.out *.sha256 main code-quality-report.json $(PKG)
	@docker rm --force bridgr_yum bridgr_python bridgr_ruby &> /dev/null || true

generate: $(GO_FILES)
	@GOOS="" go generate ./...

$(PROJECT_NAME): generate $(GO_FILES)
	@go build -tags dist -i -v -o $@ -ldflags="-X bridgr.Version=${VERSION}" $(CMD)

%.sha256:
	@openssl dgst -sha256 -hex $* | cut -f2 -d' ' > $@

download: generate
	@go mod download
