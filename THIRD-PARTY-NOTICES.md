# Third-Party Notices

This document lists third-party Go modules currently included in the
repository dependency graph through `go.mod`, along with trademark
attributions for third-party product names referenced in distributed
documentation and release materials.

## Scope

- Listed items cover modules explicitly present in `go.mod`, including indirect
  entries.
- The Go standard library is not listed here.
- Project-provided non-code assets and optional external interoperability are
  described in separate sections below.
- Additional transitive dependencies that are not represented in `go.mod` are
  reviewed during dependency updates and release validation, but are not listed
  separately by default.

## Trademark notices

- Open source license terms listed below do not grant trademark rights beyond
  customary descriptive use.
- Third-party names in this repository and its release archives are used only to
  identify compatible platforms, services, or tooling and do not imply
  affiliation, sponsorship, or endorsement.
- GitHub is a trademark of GitHub, Inc.
- Go is a trademark of Google.
- Mac and macOS are trademarks of Apple Inc., registered in the U.S. and other
  countries and regions.
- Windows is a trademark of the Microsoft group of companies.
- Linux is the registered trademark of Linus Torvalds in the U.S. and other
  countries.
- Other company, product, and service names mentioned may be trademarks of
  their respective owners.

## Optional external interoperability

### VOICEVOX

- `mcp-notify` can optionally call a separately installed and running VOICEVOX
  Engine over its HTTP API.
- VOICEVOX Engine is dual-licensed under LGPL v3 and a separate license that
  does not require source disclosure. See the
  [official Engine license](https://github.com/VOICEVOX/voicevox_engine/blob/master/LICENSE).
- VOICEVOX Engine, voice libraries, and character assets are not included in
  this repository's Go dependency graph and are not bundled, linked, or
  redistributed by this project. The release archives therefore do not contain
  VOICEVOX binaries or license files.
- Users are responsible for complying with the VOICEVOX software terms and the
  terms for each voice or character they use. See the
  [VOICEVOX software terms](https://voicevox.hiroshiba.jp/term/) and the
  [official character list](https://voicevox.hiroshiba.jp/).
- Credit placement guidance for announcements and device playback is available
  in the [official Q&A](https://voicevox.hiroshiba.jp/qa/).
- VOICEVOX is referenced only to identify compatible external software and no
  affiliation or endorsement is implied.

If a future release bundles VOICEVOX Engine, a voice library, character assets,
or generated voice samples, maintainers must reassess the distribution terms
and add all required licenses, notices, source-offer information, and credits.

## Bundled project assets

The following project-provided assets are distributed under this repository's
MIT license and have no separate third-party notice:

- `docs/assets/mcp-notify-icon.png`
- `sounds/complete.wav`
- `sounds/作業終了.wav`
- `sounds/通知 音 完了.wav`
- `sounds/alerts/sample.mp3`

Contributors must not add third-party media or VOICEVOX-generated audio to a
release without documenting its source, applicable license or terms, and any
required attribution in this file.

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
- Re-check trademark attribution needs when distributed documentation, release
  contents, or third-party tooling references change.
- Record provenance and distribution terms whenever bundled media assets are
  added or replaced.
- If a module ships multiple notices or mixed-license files, summarize that
  fact here and retain the upstream notice requirements in distributed
  materials when applicable.
