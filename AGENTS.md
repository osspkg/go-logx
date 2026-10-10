# Repository instructions

## Project

- This repository is the Go module `go.osspkg.com/logx` (Go 1.26).
- The package is at the repository root. The logger supports JSON and key/value
  formatters and adapters for `log/slog`.
- Keep the public API and existing output semantics stable unless the task asks
  for a change. Update tests and README examples when behavior changes.

## Commands

Run commands from the repository root.

- Run the test suite with `go test ./...`.
- Run the full project checks on Linux with `make ci`. This installs
  `go.osspkg.com/goppy/v3/cmd/goppy@latest` and then runs license, lint, test,
  and build checks.
- Run formatter benchmarks with, for example,
  `go test -run '^$' -bench '^(BenchmarkNewJSON|BenchmarkNewString)$' -benchmem`.

GitHub Actions runs `make ci` on Linux and `go test ./...` on macOS and Windows.
The `.golangci.yml` file defines the lint and formatting rules.

## Code changes

- Logger options are applied in `New` and the slog adapter constructors; avoid
  adding runtime-mutated configuration without accounting for concurrent use.
- Built-in formatters are implemented in `formatter_json.go` and
  `formatter_strings.go`. Keep their escaping, context handling, and line
  endings covered when changing output behavior.
- `Formatter` is the extension point for custom output. Preserve error
  propagation and writer behavior when changing formatter integration.
- Keep performance changes evidence-based. Compare equivalent benchmarks for
  elapsed time, bytes, and allocations, and retain the simpler implementation
  when a change does not produce a repeatable improvement.
