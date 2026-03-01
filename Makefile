VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X github.com/elliottmoos/clover/cmd.version=$(VERSION) \
           -X github.com/elliottmoos/clover/cmd.commit=$(COMMIT) \
           -X github.com/elliottmoos/clover/cmd.date=$(DATE)

.PHONY: build test lint clean

build:
	go build -ldflags "$(LDFLAGS)" -o clover .

test:
	go test ./... -count=1

lint:
	go vet ./...

clean:
	rm -f clover coverage.out
