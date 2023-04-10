<!--
SPDX-FileCopyrightText: 2021 Comcast Cable Communications Management, LLC
SPDX-License-Identifier: Apache-2.0
-->
# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [3.0.1]
### Changed
- Write to GITHUB_ENV instead of set-output.

## [3.0.0]
### Changed
- The external interface uses a release body file instead of a block of text.

## [2.0.2]
### Fixed
- Change the descriptions to not use nested ' characters.

## [2.0.1]
### Fixed
- Change the descriptions to not attempt to use the variable expansion tooling.

## [2.0.0]
### Changed
- Change from using the bash script to a go program that is able to error check
  better each step.

## [1.0.0]
### Added
- Initial script and action file.
- Sign everything that is in the artifacts directory so binaries could be included.

## [0.0.0]
### Added
- Initial creation

[Unreleased]: https://github.com/xmidt-org/release-builder-action/compare/v3.0.1...HEAD
[3.0.0]: https://github.com/xmidt-org/release-builder-action/compare/v3.0.0...v3.0.1
[2.0.2]: https://github.com/xmidt-org/release-builder-action/compare/v2.0.2...v3.0.0
[2.0.2]: https://github.com/xmidt-org/release-builder-action/compare/v2.0.1...v2.0.2
[2.0.1]: https://github.com/xmidt-org/release-builder-action/compare/v2.0.0...v2.0.1
[2.0.0]: https://github.com/xmidt-org/release-builder-action/compare/v1.0.0...v2.0.0
[1.0.0]: https://github.com/xmidt-org/release-builder-action/compare/v0.0.0...v1.0.0
