# Add `--server-name` and `--tool-prefix`

## Summary

- allow overriding `serverInfo.name` from the CLI
- allow prefixing the exposed MCP tool name from the CLI

## Why

- multiple registrations of the same binary are already a supported usage pattern
- some MCP clients display `serverInfo.name` in logs or UI, so a fixed name is less useful
- some clients flatten or otherwise surface tool names in a way where a configurable prefix helps avoid ambiguity

## Done when

- startup config supports `--server-name`
- startup config supports `--tool-prefix`
- `initialize` returns the configured `serverInfo.name`
- `tools/list` returns the prefixed tool name when configured
- tests and docs cover the new options
