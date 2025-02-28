NAME=elgato

default: build

all: platforms

fmt:
	go fmt .

tidy:   fmt
	go mod tidy

VERSION=$(shell git describe --tags --always --dirty)
COMMIT=$(shell git rev-parse --short HEAD)
BUILDDATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

build: tidy
	go build -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILDDATE)" -o $(NAME) .

clean:
	rm -rf $(NAME) bin index.html

linux-arm64: tidy
	GOOS=linux GOARCH=arm64  go build -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILDDATE)" -o bin/$(NAME).linux-arm64 .

linux-amd64: tidy
	GOOS=linux GOARCH=amd64  go build -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILDDATE)" -o bin/$(NAME).linux-amd64 .

darwin-arm64: tidy
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILDDATE)" -o bin/$(NAME).mac-arm64 .

darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILDDATE)" -o bin/$(NAME).mac-amd64 .

mac-arm64: darwin-arm64

mac-amd64: darwin-amd64

linux: linux-arm64 linux-amd64

mac: darwin-arm64 darwin-amd64

install: build
	sudo install -p -m0755 $(NAME) /usr/local/bin

platforms: mac linux
