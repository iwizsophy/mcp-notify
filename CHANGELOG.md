# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog, and this project follows Semantic Versioning.

## [Unreleased]

### Added

- Optional `speak_text` MCP tool backed by a separately running VOICEVOX Engine
- Runtime controls for VOICEVOX speaker/style, speed, pitch, intonation, volume, and synchronous or asynchronous playback

### Changed

- CI `push` triggers now explicitly target all branches, fixing the previous
  misconfiguration where `tags-ignore` alone prevented branch pushes from
  creating runs
- Notification files and synthesized speech now share one normalized 48 kHz stereo audio output context
- Updated `actions/checkout` from v6 to v7 and `softprops/action-gh-release` from v2 to v3
- Documented VOICEVOX installation, external-runtime licensing boundaries, generated-audio terms, and bundled asset provenance
- Added client-neutral MCP setup and dynamic speech policy guidance, with concrete Codex, Claude, and VS Code examples
- Corrected the copyright holder in the canonical MIT license to match the repository identity and reference translation

## [1.1.0] - 2026-04-04

### Added

- Each release archive now bundles a Syft-generated `SBOM.spdx.json`

### Changed

- CI now runs on pushes for any branch commit except release tags, and on pull
  requests
- The release workflow now generates an SPDX JSON SBOM from the assembled
  package contents before archiving
- Third-party notices now include trademark attributions for names used in
  shipped documentation and release materials

## [1.0.0] - 2026-04-01

### Added

- OSS repository documentation and GitHub community health files
- Issue and pull request templates
- CI and release automation for GitHub Actions

### Changed

- Server metadata now reports version `1.0.0`
- Release archives now include the linked setup, policy, and contributor documentation
- Documentation examples and verification notes now match the current runtime behavior
- Clarified contributor expectations and repository support/security policies
