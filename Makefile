PROJECT_NAME := "bridgr"
PKG := "$(PROJECT_NAME)"
CMD := "cmd/bridgr/main.go"
GO_FILES := $(shell find . -name '*.go' | grep -v _test.go)

.PHONY: all coverage lint test race x2unit xunit clean generate

all: $(PROJECT_NAME)

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
	@go test -v -race ./...

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

clean:
	@rm -rf internal/app/bridgr/assets/templates.go coverage.out yum files tests.xml tests.out coverage.out main $(PKG)

generate: $(GO_FILES)
	@go generate ./...

$(PROJECT_NAME): generate $(GO_FILES)
	@go build -tags dist -i -v -o $@ $(CMD) 

# need something in here to check $TRAVIS_TAG an add version to the build command with -X Version=${TRAVIS_TAG}
