# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

Changed:

- central OS/Architecture name handling

## [v0.6.0] - Feb 27, 2022

Changed:

- Separate Arch, ArchName, and ArmVersion, to allow sensulib modules to accurately
  filter for ARM version in sensu asset filters

## [v0.5.0] - Feb 22, 2022

Added:

- ARM 32bit support (setting with GOARM)

Changed:

- Updated dependencies, go version up to 1.17
- Change default branch to "main"

Fixed:

- read from default token file. Fixes [#9](https://github.com/julian7/goshipdone/issues/9)

## [v0.4.1] - Dec 13, 2019

Fixed:

- publish:artifact ship context missing

## [v0.4.0] - Dec 11, 2019

Added:

- documentation in README
- publish:artifact to publish to gitlab/github
- gitlab support, closes [#4](https://github.com/julian7/goshipdone/issues/4)
- search for alternative build file, closes [#1](https://github.com/julian7/goshipdone/issues/1)

Changed:

- rename setup:git_tag to setup:git
- use context.Context everywhere
- make pipeline generic, rename modules to stage
- replace environment handler with withenv
- move archive:* modules into build: stage, closes [#2](https://github.com/julian7/goshipdone/issues/2)

Fixed:

- module registration done explicitly

Removed:

- publish:gitlab and publish:github, as publish:artifact already does that

## [v0.3.0] - Nov 17, 2019

Added:

- prerelease detection

Fixed:

- save tag name for future use

## [v0.2.0] - Nov 17, 2019

Added:

- variable expansion for modules.TemplateData.Parse

Changed:

- upgrade to go-github/v28

Fixed:

- default mods overrode loaded ones
- archive:changelog to cut only the last release announcement

Removed:

- module dependencies

## [v0.1.0] - Nov 16, 2019

Added:

- archive:changelog, taking current label entry
- setup:env, with internal env.var handling
- publish:github
- publish only when defined
- publish:scp
- release_note:checksum
- archive:tar
- archive:upx
- basic structure, first services

Changed:

- rename project to goshipdone
- fold release_note into archive

[Unreleased]: https://github.com/julian7/goshipdone/compare/v0.6.0...HEAD
[v0.6.0]: https://github.com/julian7/goshipdone/compare/v0.5.0...v0.6.0
[v0.5.0]: https://github.com/julian7/goshipdone/compare/v0.4.1...v0.5.0
[v0.4.1]: https://github.com/julian7/goshipdone/compare/v0.4.0...v0.4.1
[v0.4.0]: https://github.com/julian7/goshipdone/compare/v0.3.0...v0.4.0
[v0.3.0]: https://github.com/julian7/goshipdone/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/julian7/goshipdone/compare/v0.1.0...v0.2.0
[v0.1.0]: https://github.com/julian7/goshipdone/releases/tag/v0.1.0
