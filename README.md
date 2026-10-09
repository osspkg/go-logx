# go-logx

`go-logx` is a small leveled logger for Go applications. It supports JSON and
key/value text output, plus adapters for the standard library's `log/slog`.

## Install

```sh
go get go.osspkg.com/logx
```

## Use the default logger

The default logger writes JSON to standard output and logs errors by default.
Context arguments are supplied as alternating key/value pairs; JSON context
values are encoded as strings. Context keys are JSON object fields, so their
output order is unspecified.

```go
package main

import "go.osspkg.com/logx"

func main() {
	logx.SetDefault(logx.New(logx.WithLevel(logx.LevelInfo)))
	logx.Info("server started", "address", ":8080")
}
```

## Configure a logger

Create an independent logger when a component needs its own output, level, or
formatter. The string formatter separates fields with tabs by default.

```go
package main

import "go.osspkg.com/logx"

func main() {
	logger := logx.New(
		logx.WithLevel(logx.LevelDebug),
		logx.WithFormatter(logx.NewFormatString()),
	)
	logger.Info("request handled", "status", 200, "duration", "12ms")
}
```

The `slog` adapters use the standard JSON or text handler selected by their
constructor unless `WithFormatter` is supplied, in which case that formatter
handles the output. All logger configuration is supplied when the logger is
created.

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

Call `SetDefault(logx.New(...))` during startup to configure the package-level
functions. Loggers keep their configuration after construction; concurrent
log writes are serialized so records from arbitrary writers do not interleave.

## License

BSD 3-Clause. See [LICENSE](LICENSE).
