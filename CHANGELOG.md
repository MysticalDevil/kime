# Changelog

## v1.0.4 - 2026-04-26

### What's Changed

- **feat**: Added automatic ASCII fallback rendering for non-modern terminals.
- **feat**: Added `KIME_RENDER_MODE=auto|unicode|ascii` render mode selection.
- **feat**: ASCII mode now uses ASCII-only borders, progress bars, and English labels for non-UTF-8 terminals.
- **fix**: Environment variables now override config credentials as documented.
- **fix**: Rate-limit window rendering now respects API `timeUnit` values.
- **fix**: Config writes now use a temporary file and rename for atomic updates.
- **docs**: Updated help output and project documentation for render mode, install requirements, and credential precedence.
