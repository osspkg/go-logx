# go-logx

[![Go version](https://img.shields.io/github/go-mod/go-version/osspkg/go-logx)](https://go.dev/)
[![CI](https://github.com/osspkg/go-logx/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/osspkg/go-logx/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/osspkg/go-logx)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/go.osspkg.com/logx.svg)](https://pkg.go.dev/go.osspkg.com/logx)

`go-logx` is a leveled Go logger with JSON and key/value text formatters. It also
provides adapters for the standard library's `log/slog` package.

## Contents

- [Requirements](#requirements)
- [Installation](#installation)
- [Quick start](#quick-start)
- [Configuration](#configuration)
- [Formatters](#formatters)
- [slog adapters](#slog-adapters)
- [Default logger](#default-logger)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)

## Requirements

- Go 1.26 or later.

## Installation

```sh
go get go.osspkg.com/logx
```

## Quick start

```go
package main

import (
	"os"

	"go.osspkg.com/logx"
)

func main() {
	logger := logx.New(
		logx.WithOutput(os.Stdout),
		logx.WithLevel(logx.LevelInfo),
	)

	logger.Info("server started", "address", ":8080", "port", 8080)
}
```

The default formatter is JSON. This call writes one JSON record followed by a
newline:

```json
{"time":"2026-10-10T12:00:00Z","level":"INFO","msg":"server started","ctx":{"address":":8080","port":"8080"}}
```

## Configuration

`New` accepts options for the minimum log level, output writer, and formatter.
The default level is `LevelError`, and the default writer is standard output.

Available levels, from the least restrictive to the most restrictive, are
`LevelDebug`, `LevelInfo`, `LevelWarn`, and `LevelError`. A record is written
when its level meets the configured threshold. `Fatal` writes a fatal record
and then exits the process with status 1.

Context is supplied as alternating key/value arguments:

```go
logger.Info("request handled", "method", "GET", "status", 200)
```

For the JSON formatter, context values are represented as strings. If there
are no context pairs, the `ctx` field is omitted. Context keys are JSON object
fields, so their order is unspecified.

Loggers serialize writes to their configured writer. To use a different
formatter, pass it at construction time:

```go
logger := logx.New(
	logx.WithLevel(logx.LevelDebug),
	logx.WithFormatter(logx.NewFormatString()),
)
```

## Formatters

The built-in formatters are:

- `NewFormatJSON()` writes JSON records.
- `NewFormatString()` writes quoted `key=value` fields separated by tabs.
- `NewFormatString(logx.WithDelimiter('|'))` selects another single-byte
  delimiter.

Applications can provide a custom formatter by implementing `logx.Formatter`:

```go
import (
	"io"

	"go.osspkg.com/logx"
)

type MessageOnlyFormatter struct{}

func (MessageOnlyFormatter) Encode(w io.Writer, m *logx.Message) error {
	_, err := io.WriteString(w, m.Message+"\n")
	return err
}
```

## slog adapters

Use the standard library's JSON or text handler with the same logger options:

```go
package main

import (
	"os"

	"go.osspkg.com/logx"
)

func main() {
	logger := logx.NewSLogJsonAdapter(
		logx.WithOutput(os.Stdout),
		logx.WithLevel(logx.LevelInfo),
	)
	logger.Info("worker ready", "worker", 3)
}
```

`NewSLogStringAdapter` selects `slog` text output. Passing `WithFormatter`
instead routes the adapter through the supplied `logx.Formatter`.

## Default logger

The package-level functions use a default logger initialized with JSON output
and the `LevelError` threshold. Select another logger during startup with
`SetDefault`:

```go
logx.SetDefault(logx.New(logx.WithLevel(logx.LevelInfo)))
logx.Info("application started", "version", "1.2.3")
```

Call `SetDefault` before starting concurrent logging.

## Development

Run the package tests from the repository root:

```sh
go test ./...
```

Run the full project checks with:

```sh
make ci
```

`make ci` installs the repository's `goppy` tooling and runs license, lint,
test, and build checks. GitHub Actions runs this command on Linux and runs the
Go test suite natively on macOS and Windows.

## Contributing

Report bugs or request changes through [GitHub Issues](https://github.com/osspkg/go-logx/issues).
For code changes, open a pull request and include tests for changed behavior.

## License

BSD 3-Clause. See [LICENSE](LICENSE).
