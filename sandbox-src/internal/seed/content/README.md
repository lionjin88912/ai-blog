# Seed content (build-time)

此目錄由 `make seed-sync` 在 build 前從 repo 根目錄填充(.agents/skills、GEMINI.md、HOWTO.md、BOOTSTRAP.md),
並寫入 `SEED_VERSION`。除本 README 外,內容一律不進 git。

直接 `go build`(沒跑 seed-sync)會得到「引擎版」binary — 功能正常但不帶出廠 skills。
正式散佈請一律走 `make build-all`。
