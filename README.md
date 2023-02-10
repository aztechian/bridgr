# Bridgr

> Bridging the air-gap for your artifacts

[![Gitpod Ready-to-code][gitpod-image]][gitpod-url]

[![Build Status][actions-image]][actions-url]
[![Go Version][goversion-image]][gh-gomod]
[![Release Date][releasedate-image]][gh-dl-url]
[![Release Version][releasever-image]][gh-dl-url]
[![Downloads Stats][gh-downloads]][gh-dl-url]
[![GitHub][license]][license-url]
[![Go Report Card][reportcard-image]][reportcard-url]
[![maintainability][maintainability-image]][maintainability-url]
[![coverage][coverage-image]][coverage-url]
[![codescene][codescene-image]][codescene-url]

## Use Cases

Bridr has supported the following use cases (if you have others, please let us know!):

- Simple and automated method to gather dependencies to transfer to air-gapped network
- Configuration control of build libraries and company/project dependencies
- Host a repository of custom libraries

## Introduction

Projects that need to build and/or deploy to air-gapped networks frequently run into a problem: All of the artifacts needed to build
your software (or deploy it) aren't available! The common solution to this problem has been to have each developer bring in the
pieces they need. Governance of the artifacts becomes nearly non-existant - at best you have a "dumping ground" of files that no one
person knows much about.

Bridgr helps bring order to the chaos by allowing a single, YAML manifest file to define all input artifacts to your system. With this
in place Bridgr allows:

- Version Control of artifacts - without needing the space to physically store them
- DevOps and CI Pipeline friendliness
- Software supply chain protection (reduces chance of picking up [typosquatting](https://en.wikipedia.org/wiki/Typosquatting) packages)
- Review of changes to artifacts by security teams or CM _before_ the artifact makes it to the target network
- Static website hosting of artifacts on the target network (with metadata, so repositories like YUM and Rubygems "just work")
- Support for multiple output formats - local filesystem, object storage, DVD image(?)

For more background and explanation of the use case for Bridgr, please see the [narrative](NARRATIVE.md).

For our security policy and Vulnerability Reporting Policy, please see [SECURITY](SECURITY.md).

Below is an example of Bridgr running and creating a YUM repository, downloading some files, and exporting docker images.
The file listings at the end show the artifacts created, including YUM metadata.

![header video](doc/bridgr.gif)

## Installation

Simply download the appropriate architecture binary from the [releases](releases) page, and execute it from wherever you want.

Additionally, bridgr is supported through the [asdf](https://asdf-vm.com/#/core-manage-asdf) tool. `asdf` is a version manager for many tools, languages and libraries and is highly recommended (especially if you are a software developer!).

To install with `asdf`, run

```shell
asdf plugin-add bridgr https://github.com/aztechian/asdf-bridgr.git
```

Add bridgr to your `.tool-versions` file. Then, to install bridgr to your system, run

```shell
asdf install bridgr
```

## Usage example

By default, Bridgr will create a `packages` directory with all artifacts gathered in the "current working directory" where you execute Bridgr.

Also by default, Bridgr will look for a `bridge.yaml` manifest file in the directory where it is being run. This can be overridden with the `-c` option to bridgr to specify a configuration file elsewhere.

```shell
./bridgr -c path/to/another/bridge.yml
```

To only run one of the repository types, simply give that type after any configuration options. As an example, to only run the Files type, execute Bridgr like this:

```shell
./bridgr -v files
```

### Output

Bridgr, by default will output a "spinner" display on the terminal to `stderr`. Warning-level logs will be output to `stdout`. It is possible to redirect stdout to file, an only see the spinner

```shell
bridgr > bridgr.log
```

or

```shell
bridgr > /dev/null
```

When in verbose mode, bridgr will not emit the spinner status and will only log to stdout.

Finally, when the target terminal is not a TTY (ie, when in an automated CI build) the spinner will not be shown.

### command line options

| Option              | Meaning                                                                                                                                                     |
| ------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| -v / --verbose      | Verbose Output                                                                                                                                              |
| -n / --dry-run      | Dry-run. Only do setup, don't fetch artifacts                                                                                                               |
| -c / --config       | Specify an alternate configuration file                                                                                                                     |
| --version           | Print the version of Bridgr and exit. The output of stderr can be redirected to /dev/null to get just the version string.                                   |
| -H / --host         | Run Bridgr in "hosting" mode. This mode does no downloading of artifacts, but makes Bridgr into a simple HTTP server. See `Hosting` for more detail         |
| -l / --listen       | The listen address for Bridgr in hosting mode. This is only effective when coupled with the `-H` flag. Default is `:8080`                                   |
| -x / --file-timeout | A go "duration" specifying an overall timeout for HTTP file downloads. Examples are `15s` (15 seconds), or `2h5m` (2 hours and 5 minutes). Default is `20s` |

### Artifacts requiring authentication

Bridgr supports getting authenticated artifacts for `Files`, `Docker` and `Git`. Sensitive credential information is passed to Bridgr with environment variables. It does not support putting credentials in the configuration file because it risks users comitting these credentials into version control. Bridgr intends to promote good credential hygene.

Providing credentials follows a pattern of environment variable naming

- Username -> `BRIDGR_[HOST]_USER`
- Password -> `BRIDGR_[HOST]_PASS`
- API Token -> `BRIDGR_[HOST]_TOKEN`

Only one of Password or Token can be given. If both are provided, token will override.

The `[HOST]` portion of the environment variable above should be the hostname of the URL being fetched, converted to uppercase and `.` replaced with `_`. This is most easily shown with examples.

Fetching authentication protected docker hub image:

```shell
BRIDGR_DOCKER_IO_USER=user BRIDGR_DOCKER_IO_PASS=secret bridgr docker
```

In this case we have provided a username (user) and password (secret) for the default docker registry (docker.io). When the docker worker is run, and any images are specified from docker.io, bridgr will look for these two variables for credential information.

Another example for files:

```shell
BRIDGR_PROTECTED_MYSERVER_COM_USER=user BRIDGR_PROTECTED_MYSERVER_COM_PASS=secret bridgr files
```

And, finally for git - but this time showing a token (typical with Github and Gitlab):

```shell
BRIDGR_GITHUB_COM_TOKEN=abcdefg123456789 bridgr git
```

In this case, we don't need to specify the `_USER` part of the credential, because the git worker assumes a username of `git`, and Github or Gitlab just need it to _not_ be blank. The worker does this for you.

#### S3 Authentication

For files that are specified in the bridgr configuration file that begin with `s3://`, Bridgr will use the AWS SDK to download the file from an S3 bucket source. The format of the source file must be `s3://<bucket-name>/<path>/<file>`. In other words, the bucket name must not be a DNS alias, or an HTTP location that is ultimately served by an S3 bucket - it must be the "raw" bucket name as seen in your account.

If all of the S3 file sources you wish to use are from the same account, then the usual methods of giving credentials to the S3 SDK or AWS cli [can be used](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html). We recommend also setting the environment variable `AWS_SDK_LOAD_CONFIG=1` so that all profiles in your AWS configuration may be utilized.

If you want to provide specific credentials per S3 bucket, you can use the environment variable schema as described above. In the case of S3 locations, the hostname will be the bucket name. So, for example with a bucket name of `example-bucket` you would use

`BRIDGR_EXAMPLE-BUCKET_USER` and `BRIDGR_EXAMPLE-BUCKET_PASS`

The values should be the same AWS credentials used by `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`, respectively. Bridgr will look up the environment variables for the credentials for any files in that bucket, and create a new session with those credentials to fetch the files from S3.

It is possible to download "public" files from S3 that do not require authentication. In this case, you do not need to specify credentials (if you prefer not to), as Bridgr will use anonymous credentials to access the file. This case is rare, and you would have to know the file bucket and path specifically.

## Hosting mode

Once artifacts have been gathered by Bridgr and moved across the air-gap, it is required that there be an HTTP server available on the network for serving out these artifacts. In the absense of having an existing server available, Bridgr can itself act as a simple HTTP server. When run in "hosting" mode (`-H` command line option) Bridgr will not fetch
any artifacts or look for a manifest file, but will only serve out static files from the `packages` directory where it is executed. When hosting mode is combined with the `--verbose` option, Bridgr will write HTTP logs to stdout in [Combined Log Format](http://httpd.apache.org/docs/current/logs.html#accesslog). If you desire logs be written to a file, then you are responsible for redirecting stdout to the appropriate file in your shell.

Note, that there is no complex configuration available to Bridgr in hosting mode. If you require SSL/TLS for your artifacts, then you must use another product. A containerized Nginx server would be one option, for example. Likewise, there is no authentication for artifacts in hosting mode.

However, if you need a quick-and-dirty HTTP server or as a proof-of-concept Bridgr can meet that need.

An example of running Bridgr for a long term HTTP hosting mode

```shell
nohup ./bridgr -H -v &>/var/log/bridgr &
```

You may also create a systemd service file and be able to control Bridgr as an OS service.

## Development setup

Requires Go version 1.13 or higher. Current Go version is specified in `go.mod`.

Bridgr uses Go modules available since GoLang 1.11 release. To do development on Bridgr, simply clone this repository to your preferred location
and run `make`. This will download all dependencies using the controlled go modules configuration. You must have go properly installed and configured on your system first.

Some handy make targets to help with development:

| Target   | Meaning                                                            |
| -------- | ------------------------------------------------------------------ |
| test     | Run the unit tests                                                 |
| coverage | Run the unit tests, and open a browser to show the coverage report |
| download | Only download the module dependencies                              |
| generate | Only generate the templated files to be bundled in the binary      |

The default target is to build the bridgr binary. It will create a binary named `bridgr` in the root of the repository.

### Dependencies to use

Using new (as of go 1.11) [modules-style](https://github.com/golang/go/wiki/Modules) dependencies.
Project structure following [these guidelines](https://github.com/golang-standards/project-layout)
Example project showing [CI pipeline](https://gitlab.com/pantomath-io/demo-tools)

Significant Go modules used by `Bridgr`:

- go-git
- docker.io/go-docker
- yaml.v3
- vfsgen
- helm/v3
- unknwon.dev/clog/v2

Potential for schema definition/validation of the YAML config file: [https://github.com/rjbs/rx](https://github.com/rjbs/rx)
Potential library for creating iso9660 (ISO) files [https://github.com/kdomanski/iso9660](https://github.com/kdomanski/iso9660)

## Release History

- 1.5.2
  - Add support for installing Bridgr via asdf (see Installation section)
  - Update AWS and Helm libraries
  - Fix for setting version in published binaries
  - Added Code of Conduct and Vulnerability Disclosure Policy
- 1.5.1
  - Support for Helm repository creation
  - Support downloading files from S3
  - Default verbosity now creates a "spinner" on terminal stderr. Verbose mode outputs all logging messages.
  - Migrated to [clog](https://github.com/go-clog/clog) logging library
  - Migrated to CodeClimate for coverage info and CI integration
  - Code lint issues fixed
  - Added [reviewdog](https://github.com/apps/reviewdog) to integrate PR checks with golangci-lint results
- 1.4.0
  - Complete rewrite of bridgr, to organize code internally for better testability and future work
  - Added `--file-timeout` flag to allow modifying the HTTP/s client overall timeout for downloading files
- 1.3.0
  - Add authentication support
  - Update to Go version 1.13
- 1.2.1
  - Fixes for usability bugs
- 1.2.0
  - Bridgr is now itself a static HTTP server (use the `-H` option flag)
  - Added Git repo cloning support
  - Added Rubygem repo creation
  - Built against Go version 1.12
- 1.1.0
  - Add PyPi support
- 1.0.0
  - Intial release of Bridgr with support for Yum, Files, and Docker artifacts
- 0.0.1
  - Work in progress

## Meta

Ian Martin – bridgr@imartin.io

Distributed under the MIT license. See `LICENSE` for more information.

[https://github.com/aztechian/bridgr](https://github.com/aztechian/)

## Contributing

1. [Fork it](https://github.com/aztechian/bridgr/fork)
1. Create your feature branch (`git checkout -b feature/fooBar`)
1. Commit your changes (`git commit -am 'Add some fooBar'`)
1. Push to the branch (`git push -u origin HEAD`)
1. Create a new Pull Request

<!-- Markdown link & img definitions -->

[gh-downloads]: https://img.shields.io/github/downloads/aztechian/bridgr/total.svg
[gh-dl-url]: releases/
[gh-gomod]: go.mod
[license]: https://img.shields.io/github/license/aztechian/bridgr
[license-url]: LICENSE
[actions-image]: https://img.shields.io/github/actions/workflow/status/aztechian/bridgr/ci.yaml
[actions-url]: https://github.com/aztechian/bridgr/actions/workflows/ci.yaml?query=branch%3Amaster
[reportcard-image]: https://goreportcard.com/badge/github.com/aztechian/bridgr
[reportcard-url]: https://goreportcard.com/report/github.com/aztechian/bridgr
[maintainability-image]: https://img.shields.io/codeclimate/maintainability/aztechian/bridgr?logo=code-climate
[maintainability-url]: https://codeclimate.com/github/aztechian/bridgr/maintainability
[coverage-image]: https://img.shields.io/codeclimate/coverage/aztechian/bridgr?logo=code-climate
[coverage-url]: https://codeclimate.com/github/aztechian/bridgr/test_coverage
[codescene-image]: https://codescene.io/projects/4859/status-badges/code-health
[codescene-url]: https://codescene.io/projects/4859/
[releasedate-image]: https://img.shields.io/github/release-date/aztechian/bridgr?color=blueviolet
[releasever-image]: https://img.shields.io/github/v/release/aztechian/bridgr
[gitpod-url]: https://gitpod.io/#https://github.com/aztechian/bridgr
[gitpod-image]: https://img.shields.io/badge/Gitpod-Ready--to--Code-blue?logo=gitpod
[goversion-image]: https://img.shields.io/github/go-mod/go-version/aztechian/bridgr
