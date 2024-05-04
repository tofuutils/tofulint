default: build

prepare:
	git submodule init
	git submodule update
	go run ./plugin/stub-generator

test: prepare
	go test -timeout 5m $$(go list ./... | grep -v test-fixtures | grep -v stub-generator | grep -v integrationtest)

build:
	mkdir -p dist
	go build -v -o dist/tofulint

install:
	go install

e2e: prepare install
	go test -timeout 5m ./integrationtest/...

lint:
	golangci-lint run ./...
	cd tofu/ && golangci-lint run ./...

clean:
	rm -rf dist/

generate:
	go generate ./...

release:
	go run ./tools/release/main.go

.PHONY: prepare test build install e2e lint clean generate
