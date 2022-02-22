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

## Usage

It's up to your application how to use `goshipdone`, but at some point, you want to run a pipeline:

```go
import "github.com/julian7/goshipdone"

func run() err {
    return goshipdone.Run("")
}
```

It will read `goshipdone` config from the file first found in input value, `.goshipdone.local.yml`, or `.goshipdone.yml` from the current directory (see configuration). Then, it will run the following stages:

- setup
- build
- publish (only if SKIP_PUBLISH environment variable is set to a falsey value, like "false" or "0")

It fails early, and returns an error of the first occurrence.

It is possible to register your own modules before calling `goshipdone.Run()`, which then will be available for configuration. Implement `modules.Pluggable`, and register your module with `modules.RegisterModule()`, by providing a pointer to `modules.ModuleRegistration` struct.

## Configuration

`.goshipdone.yml` file is a listing of all modules you want to run for each stage:

```yaml
---
setups:
- type: project
  name: hello_world
builds:
- type: go
  goos:
  - linux
  goarch:
  - amd64
publish:
- type: show
```

Each stage takes an array of modules, selected by their types (see below), and configured by the rest of the values.

There are automatically loaded setup modules, to provide sane default values when not defined.

## Common fields

- **id**: resulting artifact ID, other builders and publishers can take
- **skip**: OS - arch combinations to be skipped, both while building, or further handling already created artifacts. ARM (32bit) artifacts in Linux OS can have a "v5" / "v6" / "v7" suffix, reflecting to ARM v5, v6, or v7, respectively.
- **type**: module name, usually inside a stage (wrt. `*:show` as an exception)

## Default Modules

*NOTE:* module names are in `stage`:`type` format.

### *:show

No configuration.

This module is mainly for debugging purposes: it shows environment variables set, and artifacts created. This module can be loaded in every stage.

### setup:env

Default, no configuration.

This module loads environment variables into the build context. It also sets default `XDG_CONFIG_HOME` for later consumption (see `publish:artifact`).

### setup:git

Default, no configuration.

This module saves git version, current tag, current ref, and remote's URL from git information.

### setup:project

Default, parameters:

| name | default | description |
| :--- | :------ | :---------- |
| name | current directory name | Project name |
| target | dist | where to put build results |

This module defines the basic settings of the build. Project name is detected automatically by its enclosing directory (eg. name will be *hello_world* when built from `/home/rjh/projects/hello_world`).

By default, `goshipdone` will put all build artifacts into `./dist` directory, which can be overridden by `target` parameter.

### setup:skip_publish

Default, parameters:

| name | default | description |
| :--- | :------ | :---------- |
| env_name | SKIP_PUBLISH | environment variable name for instructing publish stage to be skipped |

This module reads the specified environment variable, and allows publish stage to run only, if this variable's value is falsey.

In practice, there must be a varible called SKIP_PUBLISH to be set to `false` or `0` or [any other falsey value](https://golang.org/pkg/strconv/#ParseBool).

### build:changelog

Parameters:

| name | default | description |
| :--- | :------ | :---------- |
| id   | changelog | resulting artifact ID |
| input | CHANGELOG.md | input CHANGELOG file |
| output | (empty) | output file name (== input if not specified) |

This module takes a well-formed CHANGELOG (preferred: [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)), and strips out portions for the current git tag, while keeping hyperlinks. This can then be used for release notes.

## build:checksum

Parameters:

| name | default | description |
| :--- | :------ | :---------- |
| algorithm | sha256 | checksum algo |
| builds | ["artifact"] | Array of artifacts to calculate checksum of |
| id | checksum | resulting artifact ID |
| output | {{.ProjectName}}-{{.Version}}-checsums.txt | File to write checksums to |
| skip | [] | OS - arch combinations to be skipped |

This module writes a standard checksums file using the most common algorithms (md5, sha1, sha256, sha512).

## build:go

Parameters:

| name | default | description |
| :--- | :------ | :---------- |
| after | [] | commands to run before build |
| before | [] | commands to run after build |
| goos | ["windows", "linux"] | list of GOOS values |
| goarch | ["amd64"] | list of GOARCH values |
| goarm | ["6"] | list of GOARM values (effective only if GOOS == "linux" and GOARCH == "amd64") |
| id | default | resulting artifact ID |
| ldflags | -s -w -X main.version={{.Version}} | LDFLAGS template for go build |
| main | . | module where `main()` method is defined
| output | {{.ProjectName}}{{.Ext}} | artifact file name template |
| skip | [] | OS - arch combinations to be skipped |

This module runs `go build` for each goos-goarch combination, except on skipped ones. Then it stores build result as artifact.

### build:tar

Parameters:

| name | default | description |
| :--- | :------ | :---------- |
| builds | ["artifact"] | Array of artifacts to be put into tar archives |
| commondir | {{.ProjectName}}-{{.Version}}-{{.OS}}-{{.Arch}} | topmost subdirectory name inside each tar archive |
| compression | none | compression algorithm to be used |
| files | ["README*"] | files to be copied into each tar archive |
| id | archive | resulting artifact ID |
| output | {{.ProjectName}}-{{.Version}}-{{.OS}}-{{.Arch}}.tar{{.Ext}} | artifact file name template |
| skip | [] | OS - arch combinations to be skipped |

This module takes previously built artifacts (see `builds`), and put them into a tar archive, for each OS - arch combination (except skipped ones). It is also able to put static files existing in the project directory. They will be written into archive files defined by `output` parameter, and they will be registered as an artifact identified by `id` parameter.

### build:upx

Parameters:

| name | default | description |
| :--- | :------ | :---------- |
| builds | ["default"] | Array of artifacts to be put into tar archives |
| skip | [] | OS - arch combinations to be skipped |

This module runs `upx` on each artifact file listed in `builds`, while skipping specified os-arch combinations, and replaces artifact files in place.

UPX compresses almost all kinds of executables, making them self-extracting archives. If your tool is launched infrequently, this tool can come very handy. You might not want to use it for tools invoked very frequently though; decompression uses a lot of CPU and memory.

### publish:artifact

Parameters:

| name | default | description |
| :--- | :------ | :---------- |
| builds | ["default"] | Array of artifacts to be put into tar archives |
| name | (no default) | Repository's name. No detection yet, please provide one. |
| owner | (no default) | Repository's owning organization. No detection yet, please provide one. |
| release_name | {{.Version}} | specifies the release's name |
| release_notes | (no default) | points to a noarch artifact for release notes |
| skip_tls_verify | false | disables TLS server verification. Don't use it in prod! |
| storage | github | artifact storage |
| token_env | (empty) | environment variable where auth token is specified. Autodetected when empty |
| token_file | (empty) | file name where auth token can be read from. Autodetected when empty |
| url | (empty) | artifact server's URL. Specify only for on-prem servers |

This module can publish your artifacts to a release / artifact storage server. Currently only github and gitlab are supported.

It creates a new, or edits existing release name, sets release description to the contents of `release_notes` artifact, and uploads all items of artifacts specified in `build`.

Github-specific information: token_env is `GITHUB_TOKEN`, and token_file is `$XDG_CONFIG_HOME/goshipdone/github_token`. Not tested yet on github enterprise.

Gitlab-specific information: token_env is `GITLAB_TOKEN`, and token_file is `$XDG_CONFIG_HOME/goshipdone/gitlab_token`. Specify root URL for on-prem gitlab server, `/api/v4` API will be used.

### publish:scp

Parameters:

| name | default | description |
| :--- | :------ | :---------- |
| builds | ["archive"] | Array of artifacts be uploaded |
| skip | [] | OS - arch combinations to be skipped |
| target | (empty) | SCP endpoint |

This module runs `scp` to upload builds to an SSH endpoint, using SCP. This module doesn't handle secret keys, usernames, passwords, but relies on your configuration for things like port settings, or agent usage.

## Legal

This project is licensed under [Blue Oak Model License v1.0.0](https://blueoakcouncil.org/license/1.0.0). It is not registered either at OSI or GNU, therefore GitHub is widely looking at the other direction. However, this is the license I'm most happy with: you can read and understand it with no legal degree, and there are no hidden or cryptic meanings in it.

The project is also governed with [Contributor Covenant](https://contributor-covenant.org/)'s [Code of Conduct](https://www.contributor-covenant.org/version/1/4/) in mind. I'm not copying it here, as a pledge for taking the verbatim version by the word, and we are not going to modify it in any way.

## Any issues?

Open a ticket, perhaps a pull request. We support [GitHub Flow](https://guides.github.com/introduction/flow/). You might want to [fork](https://guides.github.com/activities/forking/) this project first.
