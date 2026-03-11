# Repository Guidelines

## Project Structure & Module Organization
- `chathook-daemon/`: main Go service that receives UDP events and posts to Discord (`cmd/chathook-daemon` entrypoint, logic under `internal/chathook`).
- `udp-client/`: small Go helper binary used by BeamMP server scripts to forward payloads.
- `rsocket-module/`: optional Lua C-shared Go module (`-tags lua_module`) with unit tests.
- `Server/ChatHook/`: BeamMP Lua resource (`main.lua`, `libs/`, `config.json.sample`).
- `img/`: README assets and setup screenshots.

## Build, Test, and Development Commands
- `go build ./chathook-daemon/cmd/chathook-daemon`: build daemon locally.
- `go build -o Server/ChatHook/bin/udp ./udp-client/cmd/udp-client`: build BeamMP UDP helper.
- `go build -tags lua_module -buildmode=c-shared -o Server/ChatHook/rsocket.so ./rsocket-module`: build Linux Lua module.
- `go test ./...`: run all Go unit tests.
- `docker compose -f .docker/compose.yaml up -d`: run daemon via Docker Compose.

## Coding Style & Naming Conventions
- Go code must be `gofmt`-formatted (tabs, standard import grouping). Run `gofmt -w` on changed Go files.
- Keep package layout idiomatic: CLI in `cmd/`, non-exported app logic in `internal/`.
- Use descriptive, lower-case package names (`udpclient`, `chathook`).
- Lua files use 2-4 space indentation consistently within a file and keep helper modules in `Server/ChatHook/libs/`.

## Testing Guidelines
- Place Go tests in `*_test.go` next to implementation (examples: `core_test.go`, `service_test.go`).
- Prefer table-driven tests for config parsing and service behavior.
- Run `go test ./...` before opening a PR; add tests for new behavior and bug fixes.

## Commit & Pull Request Guidelines
- Follow Conventional Commits (`feat:`, `fix:`, `refactor:`, `docs:`). Releases are automated with release-please + GoReleaser from `main`.
- Keep commit scope focused and messages imperative, e.g. `fix: handle guest join events without API lookup`.
- PRs should include: purpose, key changes, test evidence (`go test ./...` output), and linked issues.
- For user-facing behavior changes, update `README.md` and include screenshots/log snippets when relevant.

## Security & Configuration Tips
- Never commit real webhook URLs or server secrets; use `.docker/.env.example` and `Server/ChatHook/config.json.sample` as templates.
- Validate `UDP_PORT`, bind address, and network exposure before deploying to public hosts.
