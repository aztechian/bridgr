PROJECT_NAME := "bridgr"
PKG := "$(PROJECT_NAME)"
CMD := "cmd/bridgr/main.go"
VERSION := $(shell  git describe --always --dirty | sed 's/^v//')
GO_FILES := $(shell find . -name '*.go' | grep -v _test.go)

.PHONY: all coverage lint test race x2unit xunit clean generate download

ifeq ($(GOOS), linux)
all: $(PROJECT_NAME)-Linux $(PROJECT_NAME)-Linux.sha256
else ifeq ($(GOOS), windows)
all: $(PROJECT_NAME)-Windows $(PROJECT_NAME)-Windows.sha256
else ifeq ($(GOOS), darwin)
all: $(PROJECT_NAME)-MacOS $(PROJECT_NAME)-MacOS.sha256
else
all: $(PROJECT_NAME) $(PROJECT_NAME).sha256
endif

ifeq ($(TRAVIS),)
coverage: html
else
coverage: coverage.out
endif

lint:
	@golint --set_exit_status ./...

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
	@rm -rf internal/app/bridgr/assets/templates.go coverage.out packages tests.xml tests.out coverage.out *.sha256 main $(PKG)
	@docker rm --force bridgr_yum bridgr_python bridgr_ruby &> /dev/null || true

generate: $(GO_FILES)
	@GOOS="" go generate ./...

$(PROJECT_NAME): generate $(GO_FILES)
	@go build -tags dist -i -v -o $@ -ldflags="-X main.version=${VERSION}" $(CMD)

$(PROJECT_NAME)-%: generate $(GO_FILES)
	@go build -tags dist -i -v -o $@ -ldflags="-X main.version=${VERSION}" $(CMD)
	@echo "Created executable $@"

%.sha256:
	@openssl dgst -sha256 -hex $* | cut -f2 -d' ' > $@

download: generate
	@go mod download
