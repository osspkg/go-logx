# API and configuration

The package import path is `go.osspkg.com/logx`.

## Constructors

`logx.New(options ...logx.Option)` returns `*logx.Log`. Options are applied
during construction:

- `logx.WithLevel(level)` sets the minimum accepted severity.
- `logx.WithOutput(writer)` selects an `io.Writer`; `nil` keeps the default.
- `logx.WithFormatter(formatter)` selects a formatter; `nil` keeps the
  constructor's default.

`logx.New` defaults to JSON output, `os.Stdout`, and `LevelError`. Levels from
least restrictive to most restrictive are `LevelDebug`, `LevelInfo`,
`LevelWarn`, and `LevelError`.

Methods on `*logx.Log` and the `logx.Logger` interface are `Debug`, `Info`,
`Warn`, `Error`, and `Fatal`. Calls take a message followed by alternating
context keys and values, for example:

```go
logger.Info("request handled", "method", "GET", "status", 200)
```

Keep the argument count even and use string keys. A missing value for an odd
argument list is treated as `nil`; non-string keys are converted to text.

## Package default

Package-level functions (`logx.Info`, `logx.Warn`, `logx.Error`, `logx.Debug`,
and `logx.Fatal`) use a default logger initialized with JSON output at
`LevelError`. Call `logx.SetDefault` during startup and before concurrent
logging. `SetDefault(nil)` leaves the current logger unchanged.

## slog adapters

- `logx.NewSLogJsonAdapter(options...)` uses `log/slog`'s JSON handler by
  default.
- `logx.NewSLogStringAdapter(options...)` uses `log/slog`'s text handler by
  default.

The adapters default to `LevelDebug`. If `WithFormatter` is supplied, the
adapter uses that `logx.Formatter` instead of a standard slog handler.

## Lifecycle and errors

- `Fatal` emits a fatal record and then calls `os.Exit(1)`. Deferred functions
  do not run after that call. Avoid it in reusable libraries and code that must
  clean up resources.
- Logger configuration is supplied at construction. Writes through one logger
  are serialized to its configured writer.
- The logging methods do not return write errors. A write error is printed to
  standard error. Calling a formatter's `Encode` method directly returns its
  error.
