.PHONY: build test lint lint-fix fmt tidy clean docker smoke

BINARY  := talos-mcp
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)"

build:
	CGO_ENABLED=0 go build $(LDFLAGS) -trimpath -o bin/$(BINARY) ./cmd/$(BINARY)

test:
	CGO_ENABLED=1 go test -race -count=1 -timeout 120s ./...

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

fmt:
	golangci-lint fmt ./...

tidy:
	go mod tidy

clean:
	rm -rf bin/ dist/

docker:
	docker build \
	  --build-arg VERSION=$(VERSION) \
	  --build-arg GIT_COMMIT=$(COMMIT) \
	  -t talos-mcp:$(VERSION) .

smoke: build
	@printf '%s\n%s\n%s\n' \
	  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"smoke","version":"0"}}}' \
	  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
	  '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' \
	  | bin/$(BINARY) 2>/dev/null \
	  | python3 -c "import sys,json; o=[json.loads(l) for l in sys.stdin]; ts=next(x['result']['tools'] for x in o if x.get('id')==2); print('tools:', len(ts))"
