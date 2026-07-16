.PHONY: build run clean test build-all docker-build seed-sync seed-clean icon icon-check

BINARY=ai-blog
ifeq ($(OS),Windows_NT)
	BINARY=ai-blog.exe
endif

REPO_ROOT=..
SEED_DIR=internal/seed/content

# Regenerate the Windows exe icon/version resource from build/icon.svg.
# Requires macOS (sips) + goversioninfo. The resulting .syso is committed so a
# plain `make build-all` picks up the icon without this toolchain.
#   To change the icon: edit build/icon.svg, run `make icon`, then build-all.
icon:
	@test -x build/goversioninfo || GOBIN=$(PWD)/build go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
	sips -z 512 512 -s format png build/icon.svg --out build/icon-master.png >/dev/null
	python3 build/mkico.py build
	./build/goversioninfo -64 -o resource_windows_amd64.syso build/versioninfo.json
	@rm -f build/icon-master.png
	@echo "icon resource regenerated (resource_windows_amd64.syso)"

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
	rm -f ai-blog ai-blog.exe
	rm -rf dist/

test:
	go test ./...

# Warn (don't fail) if the committed Windows icon resource is older than the
# icon source, so an edited icon.svg isn't silently shipped with the old icon.
icon-check:
	@if [ build/icon.svg -nt resource_windows_amd64.syso ]; then \
		echo "⚠️  build/icon.svg 比 resource_windows_amd64.syso 新 — 跑 'make icon' 重生成圖示再 build"; \
	fi

build-all: icon-check seed-sync
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o dist/ai-blog-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o dist/ai-blog-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o dist/ai-blog-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o dist/ai-blog-windows-amd64.exe .
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o dist/pwsh.exe ./cmd/pwsh-shim

docker-build:
	mkdir -p dist
	docker compose run --rm build
