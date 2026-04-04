# Third-Party Notices

This document lists third-party Go modules currently included in the
repository dependency graph through `go.mod`.

## Scope

- Listed items cover modules explicitly present in `go.mod`, including indirect
  entries.
- The Go standard library is not listed here.
- Additional transitive dependencies that are not represented in `go.mod` are
  reviewed during dependency updates and release validation, but are not listed
  separately by default.

## Current modules

### github.com/ebitengine/oto/v3 v3.4.0

- License: Apache License 2.0
- Source: `github.com/ebitengine/oto/v3`

### github.com/go-audio/audio v1.0.0

- License: Apache License 2.0
- Source: `github.com/go-audio/audio`

### github.com/go-audio/wav v1.1.0

- License: Apache License 2.0
- Source: `github.com/go-audio/wav`

### github.com/hajimehoshi/go-mp3 v0.3.4

- License: Apache License 2.0
- Source: `github.com/hajimehoshi/go-mp3`

### github.com/ebitengine/purego v0.9.0

- License: Apache License 2.0
- Source: `github.com/ebitengine/purego`

### github.com/go-audio/riff v1.0.0

- License: Apache License 2.0
- Source: `github.com/go-audio/riff`

### golang.org/x/sys v0.36.0

- License: BSD 3-Clause
- Source: `golang.org/x/sys`

## Update policy

- Update this file when a dependency is added, removed, or its version changes
  in `go.mod`.
- Re-check license terms when dependency versions change.
- If a module ships multiple notices or mixed-license files, summarize that
  fact here and retain the upstream notice requirements in distributed
  materials when applicable.
