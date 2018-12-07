# Change Log

All notable changes to this project will be documented in this file.

The format follows [keepachangelog.com]. Please stick to it.

## [unreleased]

Changes that are not yet released can be prepared here.

### Fixed

### Added

### Changed

### Removed

### Deprecated

## [0.3.0 Galloping Galapgos] -- 2018-12-07

### Fixed

- Compression guessing is now using Go's http.DetectContentType()

### Added

* New gateway subcommand and feature. Now files and directories can be easily
  shared to non-brig users via a normal webserver. Also includes easy https setup.

### Changed

### Removed

### Deprecated

## [0.2.0 Baffling Buck] -- 2018-11-21

### Fixed

All features mentioned in the documentation should work now.

### Added

Many new features, including password management, partial diffs and partial syncing.

### Changed

Many internal things. Too many to list in this early stage.

### Removed

Nothing substantial.

### Deprecated

Nothing.

## [0.1.0 Analphabetic Antelope] -- 2018-04-21

Initial release on the Linux Info Day 2018 in Augsburg.

[unreleased]: https://github.com/sahib/rmlint/compare/master...develop
[0.1.0]: https://github.com/sahib/brig/releases/tag/v0.1.0
[keepachangelog.com]: http://keepachangelog.com/
