.PHONY: build run clean test build-all docker-build

BINARY=ai-sandbox
ifeq ($(OS),Windows_NT)
	BINARY=ai-sandbox.exe
endif

build:
	go build -ldflags "-s -w" -o $(BINARY) .
ifeq ($(OS),Windows_NT)
	go build -ldflags "-s -w" -o sandbox/bin/pwsh.exe ./cmd/pwsh-shim
endif

run: build
	./$(BINARY)

clean:
	rm -f ai-sandbox ai-sandbox.exe
	rm -rf dist/

test:
	go test ./...

build-all:
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o dist/ai-sandbox-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o dist/ai-sandbox-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o dist/ai-sandbox-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o dist/ai-sandbox-windows-amd64.exe .
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o dist/pwsh.exe ./cmd/pwsh-shim

docker-build:
	mkdir -p dist
	docker compose run --rm build
