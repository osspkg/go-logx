---
name: go-logx-usage
description: >-
  Use the go.osspkg.com/logx Go logging library in application code. Apply when
  a task asks how to configure loggers, select formatters or slog adapters, or
  write correct go-logx examples. Use Go implementation skills for changes to
  the logger library itself.
---

# Using go-logx

Use the public package `go.osspkg.com/logx`. Confirm the version in the
consumer's `go.mod` and check the API available in that version before relying
on repository-specific details.

Choose only the reference needed for the task:

- [API and configuration](references/api.md) for constructors, options, levels,
  defaults, and lifecycle behavior.
- [Output formats](references/formats.md) for JSON, string, and slog output
  semantics.
- [Code examples](references/examples.md) for complete setup and integration
  snippets.

Prefer constructor options (`WithLevel`, `WithOutput`, `WithFormatter`) so
configuration is set before logging starts. Set the package default with
`SetDefault` during startup, before concurrent logging. Use key/value pairs for
context and keep keys as strings. Account for JSON context values being encoded
as strings. Do not call `Fatal` from reusable library code: it terminates the
process with `os.Exit(1)`.

When editing a consumer project, follow its existing logger ownership and
configuration conventions. Avoid changing the library's public API to solve a
usage problem unless the user asks for a library change.
