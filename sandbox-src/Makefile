.PHONY: build run clean test build-all docker-build seed-sync seed-clean

BINARY=ai-sandbox
ifeq ($(OS),Windows_NT)
	BINARY=ai-sandbox.exe
endif

REPO_ROOT=..
SEED_DIR=internal/seed/content

# Populate the embedded seed from the repo root.
# Excludes per-machine user state: credentials, publish records, drafts.
seed-sync:
	rm -rf $(SEED_DIR)/.agents
	mkdir -p $(SEED_DIR)/.agents/skills
	rsync -a \
		--exclude 'wp-config.json' \
		--exclude 'published.json' \
		--exclude 'articles/' \
		--exclude '__pycache__/' \
		--exclude '.DS_Store' \
		$(REPO_ROOT)/.agents/skills/ $(SEED_DIR)/.agents/skills/
	cp $(REPO_ROOT)/GEMINI.md $(REPO_ROOT)/HOWTO.md $(REPO_ROOT)/BOOTSTRAP.md $(SEED_DIR)/
	printf '%s\n' "$$(date +%Y%m%d-%H%M%S)-$$(git -C $(REPO_ROOT) rev-parse --short HEAD 2>/dev/null || echo nogit)" > $(SEED_DIR)/SEED_VERSION
	@echo "Seed synced: $$(cat $(SEED_DIR)/SEED_VERSION)"

seed-clean:
	rm -rf $(SEED_DIR)/.agents $(SEED_DIR)/GEMINI.md $(SEED_DIR)/HOWTO.md $(SEED_DIR)/BOOTSTRAP.md $(SEED_DIR)/SEED_VERSION

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
