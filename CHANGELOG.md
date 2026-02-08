# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `/release` skill now runs `/docs-doctor` validation before release (hard blocker)
- Pre-commit hook enforces CHANGELOG.md changes on feature branches
- Post-merge hook runs `orc doctor` on master/main branch
- `/docs-doctor` skill checks for repo-agnosticism violations in skills
- Guardrail enforcement documentation in CLAUDE.md
- Self-test skill now verifies tmux session management

### Changed

### Deprecated

### Removed

### Fixed

### Security
