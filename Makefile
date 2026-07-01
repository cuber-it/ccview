BIN     := ccview
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

# deploy target: install location and systemd --user service (override as needed)
PREFIX  ?= $(HOME)/.local
BINDIR  := $(PREFIX)/bin
SERVICE ?= ccview

.PHONY: all build test vet race clean cross run deploy

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

# deploy: build with the version stamp straight into BINDIR, then atomically
# replace the installed binary and restart the service. The build+mv avoids
# ETXTBSY: overwriting the running binary in place (cp) fails, but building a
# sibling and renaming over it works (the running process keeps the old inode).
deploy: vet test
	@mkdir -p $(BINDIR)
	go build -ldflags "$(LDFLAGS)" -o $(BINDIR)/$(BIN).new ./cmd/ccview
	mv -f $(BINDIR)/$(BIN).new $(BINDIR)/$(BIN)
	systemctl --user restart $(SERVICE)
	@systemctl --user is-active $(SERVICE) && echo "deployed $(VERSION) → $(BINDIR)/$(BIN)"
