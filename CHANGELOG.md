# Changelog

## v1.1.4 - 2026-04-26

### What's Changed

- **fix**: Config atomic writes now clean up the temporary file when `os.Rename` fails.
- **fix**: Render mode detection now uses `os.Getenv`/`os.Environ` directly instead of custom wrapper functions.
- **fix**: `modeUsesUnicode` now checks an explicit `RenderMode` field instead of inferring from the progress-bar character.
- **refactor**: `displayText` now reuses a package-level `strings.Replacer` for ASCII emoji stripping.
- **test**: Expanded `TestShouldUseASCII` and `TestResolveRenderMode` with edge cases for empty TERM, unknown terminal, `LC_ALL` precedence, and invalid env fallback.
- **docs**: Removed duplicate usage example in `README_zh.md`.

## v1.1.3 - 2026-04-26

### What's Changed

- **feat**: Added automatic ASCII fallback rendering for non-modern terminals.
- **feat**: Added `KIME_RENDER_MODE=auto|unicode|ascii` render mode selection.
- **feat**: ASCII mode now uses ASCII-only borders, progress bars, and English labels for non-UTF-8 terminals.
- **fix**: Config writes now use a temporary file and rename for atomic updates.
- **docs**: Updated help output and project documentation for render mode, install requirements, and credential precedence.
- **docs**: Backfilled `CHANGELOG.md` with historical release notes from `v1.0.0` through `v1.1.2`.

## v1.1.2 - 2026-04-18

### What's Changed

- **chore**: add ast-grep rule to detect passthrough wrappers. The rule is configured in
  `rules/passthrough-wrapper.yml` and can be run via `sg scan`.
- **docs**: add missing godoc comments to unexported helpers in cache, config, and main
  packages. Unified the duplicate package comments in the api package into a single
  comment in `client.go`.

## v1.1.1 - 2026-04-18

### What's Changed

- **fix**: environment variables (`KIME_TOKEN`, `KIME_DEVICE_ID`, `KIME_SESSION_ID`,
  `KIME_USER_ID`) now correctly override `config.json` values as documented. Previously,
  config values always took precedence, breaking temporary credential switching via env vars.
- **fix**: reset times under 1 hour are now shown as "N minutes later" instead of being
  truncated to "0 hours later", eliminating the misleading impression that the quota had
  already reset.
- **fix**: `selectPrimaryBalance` now checks `ExpireTime` before picking `FEATURE_OMNI` or
  `FEATURE_CODING`. If the highest-priority balance has expired, the UI falls back to the
  next valid balance instead of displaying stale data.
- **style**: resolved `golangci-lint` wsl blank-line warnings across modified files.

## v1.1.0 - 2026-04-18

### What's Changed

- **fix**: mock mode now bypasses credential resolution in `NewClient`, so setting
  `KIME_MOCK=1` lets `kime check` run as documented without requiring auth environment
  variables or a saved config.
- **test**: added coverage to ensure mock-mode client creation works without credentials and
  updated the CLI mock-mode test to verify the documented zero-credential path.

## v1.0.5 - 2026-04-18

### What's Changed

- **test**: Expanded API, cache, config, UI, and CLI coverage. Added embedded mock payload
  fixtures under `api/testdata`, `httptest` coverage for the API client, and subprocess
  coverage for CLI entry paths. Total coverage is now gated at **73%** and currently sits
  at **73.9%**.
- **ci**: Added a repository `justfile` and unified formatting, lint, test, and coverage
  commands behind `just`. GitHub Actions now use the same entrypoints, and the test workflow
  trigger matches `fmt`/`lint` on `push` to `main` and `pull_request` against `main`.
- **docs**: Added `AGENTS.md` as a contributor guide covering repository structure, core
  commands, style expectations, testing, and release hygiene.
- **chore**: Ignore local `.codex` workspace state files.

## v1.0.4 - 2026-04-17

### What's Changed

- **fix**: `loadSubscription` now tries cache before the live network request. If the
  subscription API is temporarily unreachable, the CLI falls back to valid cached data
  instead of failing entirely. `KIME_FORCE_REFRESH=1` still bypasses cache as documented.
- **fix**: `Render` no longer panics when `usages` is `nil`.
- **fix**: `selectPrimaryBalance` now picks the correct balance by feature
  (`FEATURE_OMNI` > `FEATURE_CODING` > non-expired) instead of blindly taking
  `balances[0]`.
- **fix**: `formatWindow` respects `LimitWindow.TimeUnit` (second / minute / hour / day)
  instead of assuming everything is minutes.
- **feat**: Windows support - `config` and `cache` now use platform-appropriate directories on Windows.
- **docs**: Added Windows PowerShell usage notes to README.
- **docs**: Added API reference with request examples.
- **ci**: Added GitHub Actions workflows for formatting, linting, and cross-platform testing.

## v1.0.3 - 2026-04-15

### What's Changed

- **feat**: Redesigned help output with Lipgloss-styled sections, rounded borders, and soft-green accents.
- Help now uses a structured card layout with clearer alignment for commands, flags, and environment variables.

## v1.0.2 - 2026-04-15

### What's Changed

- **fix**: Validate cached subscription data (title + end time + capabilities) before using it.
- **fix**: Use atomic file writes (tmp + rename) for cache to prevent corruption on interrupt.
- Fallback to live API data and rewrite cache when cached data is invalid or expired.

## v1.0.1 - 2026-04-15

### What's Changed

- **feat**: Usage ratio now renders with a green-to-red gradient.
- **feat**: Subscription metadata (plan, validity, capabilities) is cached until
  `currentEndTime` instead of a fixed TTL.
- **feat**: Added `KIME_FORCE_REFRESH` env var to force a full cache refresh.
- **docs**: Updated README to reflect new caching behavior and environment variables.

## v1.0.0 - 2026-04-13

First stable release.

Highlights:

- Subcommand-based CLI: `check` to fetch stats, `init` for interactive setup, default shows help.
- Full test coverage across all packages.
- Multilingual support: `zh`, `zh_TW`, `en`, `ja`.
- Built with Go 1.26+ and `encoding/json/v2`.
