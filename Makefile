VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -ldflags "-X github.com/tomo/searxng-cli/cmd.Version=$(VERSION) -X github.com/tomo/searxng-cli/cmd.Commit=$(COMMIT) -X github.com/tomo/searxng-cli/cmd.Date=$(DATE)"

.PHONY: all build test clean

all: clean build test

build:
	go build $(LDFLAGS) -o searxng-cli .

test:
	go test -v ./...

clean:
	rm -f searxng-cli

release:
	goreleaser release --clean

release-snapshot:
	goreleaser release --snapshot --clean
