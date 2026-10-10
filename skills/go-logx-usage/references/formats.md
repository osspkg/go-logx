# Output formats

## JSON formatter

`logx.NewFormatJSON()` writes one JSON object and a newline. The record contains
`time`, `level`, and `msg`; it includes `ctx` only when context is present.
Time uses RFC3339Nano.

Context values are converted to text and stored as JSON strings. For example,
an integer context value `200` appears as `"200"`, and `nil` appears as
`"null"`. Context keys become JSON object fields; their order is unspecified.

Use JSON output when records need a machine-readable outer structure. Do not
assume that context values retain their original JSON types.

## String formatter

`logx.NewFormatString()` writes quoted key/value fields separated by tabs, then
a newline. Select a different single-byte delimiter with
`logx.NewFormatString(logx.WithDelimiter('|'))`.

The format is intended for readable log lines, not as a general-purpose
serialization format. If downstream consumers parse it, preserve the chosen
delimiter and escaping expectations.

## slog output

Without a custom formatter, the adapters delegate to Go's `log/slog` JSON or
text handler. Their encoding details follow the standard library handler.
Supplying `logx.WithFormatter(...)` switches the adapter to the selected
`go-logx` formatter and its output semantics.
