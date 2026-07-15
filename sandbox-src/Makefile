.PHONY: build run clean test build-all docker-build seed-sync seed-clean

BINARY=ai-sandbox
ifeq ($(OS),Windows_NT)
	BINARY=ai-sandbox.exe
endif

REPO_ROOT=..
SEED_DIR=internal/seed/content

# Populate the embedded seed from the repo root via a cross-platform Go tool
# (works on Windows; no rsync/cp/printf). User state and untracked local
# personas are filtered by the tool, sharing seed.IsUserState as the policy.
seed-sync:
	go run ./tools/seedsync $(REPO_ROOT) $(SEED_DIR)

seed-clean:
	go run ./tools/seedsync --clean $(REPO_ROOT) $(SEED_DIR) 2>/dev/null || true

build: seed-sync
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

build-all: seed-sync
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o dist/ai-sandbox-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o dist/ai-sandbox-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o dist/ai-sandbox-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o dist/ai-sandbox-windows-amd64.exe .
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o dist/pwsh.exe ./cmd/pwsh-shim

docker-build:
	mkdir -p dist
	docker compose run --rm build
