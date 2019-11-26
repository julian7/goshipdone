# Go Ship Done

[![GoDoc](https://godoc.org/github.com/julian7/goshipdone?status.svg)](https://godoc.org/github.com/julian7/goshipdone)

This project aims to provide basic release functionality to any project. Originally it was prepared to be used with [magefile](https://magefile.org/), but the service is generic enough to allow any kinds of use.

PLEASE NOTE this project is still in early stages.

## Goals

- overridable build configuration
- provide multi-target, multi-OS, multiarch builds
- build results post-processing
- build artifacts based on builds
- signature, checksum, assetfile generator for artifacts
- artifact upload (gitlab, github, gitea, possibly bitbucket)

## Other goals

There are other goals on the horizon, which are not immediately important:

- package generator (NPFM, scoop, homebrew)
- docker builder (kaniko) as an archival tool

## Try it

Running `go run build/build.go` takes example .goshipdone.yml file, and runs it. Now it takes an optional argument, `-publish`, which enables publishing stage.

## Legal

This project is licensed under [Blue Oak Model License v1.0.0](https://blueoakcouncil.org/license/1.0.0). It is not registered either at OSI or GNU, therefore GitHub is widely looking at the other direction. However, this is the license I'm most happy with: you can read and understand it with no legal degree, and there are no hidden or cryptic meanings in it.

The project is also governed with [Contributor Covenant](https://contributor-covenant.org/)'s [Code of Conduct](https://www.contributor-covenant.org/version/1/4/) in mind. I'm not copying it here, as a pledge for taking the verbatim version by the word, and we are not going to modify it in any way.

## Any issues?

Open a ticket, perhaps a pull request. We support [GitHub Flow](https://guides.github.com/introduction/flow/). You might want to [fork](https://guides.github.com/activities/forking/) this project first.
