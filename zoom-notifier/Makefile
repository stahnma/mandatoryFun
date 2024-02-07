

NAME=zoom-notifier

default: build

all: platforms

fmt:
	go fmt .

tidy:   fmt
	go mod tidy

build:  tidy
	go build .

clean:
	rm -rf $(NAME) bin

linux-arm64: tidy
	GOOS=linux GOARCH=arm64  go build -o bin/$(NAME).linux-arm64 .

linux-amd64: tidy
	GOOS=linux GOARCH=amd64  go build -o bin/$(NAME).linux-amd64 .

darwin-arm64: tidy
	GOOS=darwin GOARCH=arm64 go build -o bin/$(NAME).mac-arm64 .

darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -o bin/$(NAME).mac-amd64 .

mac-arm64: darwin-arm64

mac-amd64: darwin-amd64

linux: linux-arm64 linux-amd64

mac: darwin-arm64 darwin-amd64

install: build
	sudo install -p -m0755 zoomwh /usr/local/bin

 
platforms: mac linux
