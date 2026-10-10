# Code examples

These examples use the public API in this repository. Import path:
`go.osspkg.com/logx`.

## Independent JSON logger

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

## String formatter

```go
logger := logx.New(
	logx.WithLevel(logx.LevelDebug),
	logx.WithFormatter(logx.NewFormatString()),
)
logger.Debug("request handled", "method", "GET", "status", 200)
```

Use a custom delimiter when the log consumer requires it:

```go
formatter := logx.NewFormatString(logx.WithDelimiter('|'))
logger := logx.New(logx.WithFormatter(formatter))
logger.Info("worker ready", "worker", 3)
```

## slog adapter

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

Use `logx.NewSLogStringAdapter` for the standard slog text handler.

## Package-level default

Select the default logger once during startup, before starting goroutines that
log through package-level functions:

```go
logx.SetDefault(logx.New(
	logx.WithLevel(logx.LevelInfo),
	logx.WithOutput(os.Stdout),
))

logx.Info("application started", "version", "1.2.3")
```

This snippet needs imports for `os` and `go.osspkg.com/logx`.
