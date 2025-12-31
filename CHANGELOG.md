# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- Updated minimum Go version to 1.22

## [0.3.0] - 2024-12-31

### Added
- Phase 3 implementation complete
- Enhanced documentation (doc.go)
- Improved test coverage (73.9%)

## [0.2.0] - 2024-12-15

### Added
- Multi-database support (PostgreSQL, MySQL, SQLite)
- Distributed locking mechanisms
- Batch tracking for precise rollback control

## [0.1.0] - 2024-12-01

### Added
- Initial release
- Database migration management
- Type-safe migrations in Go
- Transaction-wrapped execution
- Up/Down migration support
- Beautiful CLI with progress tracking
- Comprehensive logging

[Unreleased]: https://github.com/toutaio/toutago-sil-migrator/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/toutaio/toutago-sil-migrator/releases/tag/v0.3.0
[0.2.0]: https://github.com/toutaio/toutago-sil-migrator/releases/tag/v0.2.0
[0.1.0]: https://github.com/toutaio/toutago-sil-migrator/releases/tag/v0.1.0
