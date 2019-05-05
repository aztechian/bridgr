# Bridgr

> Bridging the air-gap for your artifacts

[![Build Status][travis-image]][travis-url]
[![Downloads Stats][gh-downloads]][gh-dl-url]
[![GitHub][license]][license-url]

Projects that need to build and/or deploy to air-gapped networks frequently run into a problem: All of the artifacts needed to build
your software (or deploy it) isn't available! The common solution to this problem has been to have each developer bring in the
pieces they need. Governance of the artifacts becomes nearly non-existant - at best you have a "dumping ground" of files that no one
person knows much about.

Bridgr helps bring order to the chaos by allowing a single, YAML manifest file to define all input artifacts to your system. With this
in place Bridgr can allow:

- CM control of artifacts - without necessarily needing the space to physically store them
- DevOps and pipeline friendly-ness
- Review of changes to artifacts by security teams or CM _before_ the artifact makes it to the target network
- Static website hosting of artifacts on the target network (with metadata, so repositories like YUM and Rubygems work)

![](header.png)

## Installation

OS X & Linux:

```sh
npm install my-crazy-module --save
```

Windows:

```sh
edit autoexec.bat
```

## Usage example

A few motivating and useful examples of how your product can be used. Spice this up with code blocks and potentially more screenshots.

_For more examples and usage, please refer to the [Wiki][wiki]._

## Development setup

Describe how to install all development dependencies and how to run an automated test-suite of some kind. Potentially do this for multiple platforms.

```sh
make install
npm test
```

Requires Go version 1.11 or higher.

### Dependencies to use

Using new (as of go 1.11) [modules-style](https://github.com/golang/go/wiki/Modules) dependencies.
Project structure following [these guidelines](https://github.com/golang-standards/project-layout)

We will use the following libraries to do heavy lifting:

- go-git
- docker.io/go-docker (also compare with github.com/fsouza/go-dockerclient)

Potential for schema definition/validation of the YAML config file: https://github.com/rjbs/rx

## Release History

- 0.0.1
  - Work in progress

## Meta

Ian Martin – ian@imartin.net

Distributed under the MIT license. See ``LICENSE`` for more information.

[https://github.com/aztechian/bridgr](https://github.com/aztechian/)

## Contributing

1. Fork it (<https://github.com/aztechian/bridgr/fork>)
2. Create your feature branch (`git checkout -b feature/fooBar`)
3. Commit your changes (`git commit -am 'Add some fooBar'`)
4. Push to the branch (`git push -u origin feature/fooBar`)
5. Create a new Pull Request

<!-- Markdown link & img dfn's -->
[gh-downloads]: https://img.shields.io/github/downloads/aztechian/bridgr/total.svg
[gh-dl-url]: releases/
[license]: https://img.shields.io/github/license/aztechian/bridgr.svg
[license-url]: LICENSE
[travis-image]: https://img.shields.io/travis/aztechian/bridgr/master.svg?style=flat-square
[travis-url]: https://travis-ci.org/aztechian/bridgr
[wiki]: https://github.com/aztechian/bridgr/wiki
