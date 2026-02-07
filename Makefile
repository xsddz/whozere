.PHONY: build build-all test clean

VERSION := 0.1.0
BINARY := whozere
LDFLAGS := -s -w

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/whozere

build-all:
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-darwin-amd64 ./cmd/whozere
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-darwin-arm64 ./cmd/whozere
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-linux-amd64 ./cmd/whozere
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-linux-arm64 ./cmd/whozere
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-windows-amd64.exe ./cmd/whozere

test:
	go test -v ./...

clean:
	rm -f $(BINARY)
	rm -rf dist/

run:
	go run ./cmd/whozere

install:
	go install ./cmd/whozere
