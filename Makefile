HOSTNAME=registry.terraform.io
NAMESPACE=davispalomino
NAME=vtex
BINARY=terraform-provider-${NAME}
VERSION=0.1.0
OS_ARCH=$(shell go env GOOS)_$(shell go env GOARCH)

default: install

build:
	CGO_ENABLED=0 go build -o ${BINARY}

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

test:
	go test -v ./...

testacc:
	TF_ACC=1 go test -v ./... -timeout 120m

fmt:
	go fmt ./...

lint:
	golangci-lint run

docs:
	go generate ./...

clean:
	rm -f ${BINARY}
	rm -rf ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}

deps:
	go mod download
	go mod tidy

.PHONY: build install test testacc fmt lint docs clean deps
