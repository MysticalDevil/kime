set shell := ["bash", "-euo", "pipefail", "-c"]

min_coverage := env_var_or_default("MIN_COVERAGE", "73.0")

export GOEXPERIMENT := "jsonv2"

default:
  @just --list

fmt:
  go tool gofumpt -w .

fmt-check:
  diff="$(go tool gofumpt -d .)"; \
  if [ -n "$diff" ]; then \
    printf '%s\n' "$diff"; \
    exit 1; \
  fi

lint:
  go tool golangci-lint run --fix ./...

lint-check:
  go tool golangci-lint run ./...

test:
  go test ./...

coverage:
  go test -coverprofile=cover.out ./...
  total="$(go tool cover -func=cover.out | awk '/^total:/ { sub(/%/, "", $3); print $3 }')"; \
  awk -v total="$total" -v min="{{min_coverage}}" 'BEGIN { \
    if ((total + 0) < (min + 0)) { \
      printf("coverage %s%% is below minimum %s%%\n", total, min); \
      exit 1; \
    } \
    printf("coverage %s%% meets minimum %s%%\n", total, min); \
  }'

check: fmt-check lint-check test coverage
