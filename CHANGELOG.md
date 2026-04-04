# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog, and this project follows Semantic Versioning.

## [Unreleased]

### Changed

- CI now runs on pushes for any branch commit except release tags, and on pull
  requests

## [1.1.0] - 2026-04-04

### Added

- Each release archive now bundles a Syft-generated `SBOM.spdx.json`

### Changed

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
