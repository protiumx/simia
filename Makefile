.PHONY: build run run-wasm build-wasm test
.DEFAULT_GOAL := run

VERSION := v0.1.0
LDFLAGS=-ldflags "-X=main.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o bin/simia cmd/repl/main.go

run:
	go run cmd/repl/main.go

run-wasm: build-wasm
	@echo "Starting server"
	@python3 -m http.server 8080 -d cmd/wasm

build-wasm: cp-wasmjs
	@echo "Building wasm..."
	GOOS=js GOARCH=wasm go build $(LDFLAGS) -o ./cmd/wasm/simia.wasm ./cmd/wasm/main.go 

test:
	go test ./...

.PHONY: cp-wasmjs

cp-wasmjs:
ifeq (,$(wildcard ./cmd/wasm/js/wasm_exec.js))
	cp "$(shell go env GOROOT)/misc/wasm/wasm_exec.js" cmd/wasm/js 
endif

