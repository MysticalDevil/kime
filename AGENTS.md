# Repository Guidelines

## Project Structure & Module Organization

`kime` is a small Go CLI organized by package at the repository root. Core entrypoints live in `main.go` and `main_test.go`. API client code is in `api/`, cached subscription storage is in `cache/`, user config handling is in `config/`, translations are in `i18n/`, and terminal rendering is in `ui/`. API examples and protocol notes live in `docs/API.md`. Mock API payloads used by tests and mock mode are stored in `api/testdata/`.

## Build, Test, and Development Commands

Use `just` as the default task runner.

- `just fmt`: format all Go files with `go tool gofumpt`.
- `just fmt-check`: fail if formatting changes are required.
- `just lint`: run `go tool golangci-lint run --fix ./...` for local autofixes.
- `just lint-check`: run `go tool golangci-lint run ./...` without modifying files.
- `just test`: run the full Go test suite.
- `just coverage`: generate `cover.out` and enforce the minimum total coverage.
- `just check`: run formatting, lint, tests, and coverage in one pass.

This project requires `GOEXPERIMENT=jsonv2`; the `justfile` exports it automatically. For a local build, use `GOEXPERIMENT=jsonv2 go build -o kime`.

If you need to run tools directly instead of through `just`, use:

- `go tool gofumpt -w .` for formatting
- `go tool gofumpt -d .` to check formatting without editing files
- `go tool golangci-lint run --fix ./...` for local lint autofixes
- `go tool golangci-lint run ./...` for CI-style lint checks

## Coding Style & Naming Conventions

Follow standard Go layout and keep package APIs small and explicit. Format with `go tool gofumpt`; do not hand-format. Lint with `golangci-lint` before sending changes. Use descriptive Go names: exported identifiers in `CamelCase`, unexported helpers in `camelCase`, and test functions in `TestXxx`. Keep JSON fixtures out of source when they get long; prefer files under `*/testdata/`.

## Testing Guidelines

Tests use Go’s built-in `testing` package and live alongside code in `*_test.go` files. Prefer table-driven tests for pure logic and `httptest` for API behavior. CLI paths may be tested with subprocess helpers when needed. Run `just test` locally, then `just coverage`. The repository currently enforces a minimum total coverage of `73%`; do not lower it to merge a change.

## Commit & Pull Request Guidelines

Use Conventional Commit prefixes already present in history, such as `feat:`, `fix:`, `docs:`, and `ci:`. Keep each commit scoped to one concern. Pull requests should describe the behavior change, list verification commands run, and note any config, cache, or CLI-facing effects. Include terminal output or screenshots only when UI output changes materially.

## Security & Configuration Tips

Never commit real tokens or user-specific config. Runtime config is stored in the platform config directory as `kime/config.json`; cached membership data is stored separately in the platform cache directory as `kime/membership.json`.
