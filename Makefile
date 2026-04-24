BIN     := ccview
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: all build test vet race clean cross run

all: vet test build

build:
	go build -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/ccview

test:
	go test ./... -count=1

race:
	go test ./... -count=1 -race

vet:
	go vet ./...

clean:
	rm -rf build dist $(BIN)

cross: clean
	mkdir -p dist
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BIN)-linux-amd64       ./cmd/ccview
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BIN)-linux-arm64       ./cmd/ccview
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BIN)-darwin-amd64      ./cmd/ccview
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BIN)-darwin-arm64      ./cmd/ccview
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BIN)-windows-amd64.exe ./cmd/ccview
	@ls -lh dist/

run: build
	./$(BIN)
