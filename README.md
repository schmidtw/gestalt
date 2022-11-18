<!--
SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
SPDX-License-Identifier: Apache-2.0
-->
# goschtalt
A simple configuration library that supports multiple files and formats.

[![Build Status](https://github.com/goschtalt/goschtalt/actions/workflows/ci.yml/badge.svg)](https://github.com/goschtalt/goschtalt/actions/workflows/ci.yml)
[![codecov.io](http://codecov.io/github/goschtalt/goschtalt/coverage.svg?branch=main)](http://codecov.io/github/goschtalt/goschtalt?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/goschtalt/goschtalt)](https://goreportcard.com/report/github.com/goschtalt/goschtalt)
[![GitHub Release](https://img.shields.io/github/release/goschtalt/goschtalt.svg)](https://github.com/goschtalt/goschtalt/releases)
[![GoDoc](https://pkg.go.dev/badge/github.com/goschtalt/goschtalt)](https://pkg.go.dev/github.com/goschtalt/goschtalt)

## Goals & Themes

* Favor small, simple designs.
* Keep dependencies to a minimum.
* Favor user customization options over building everything in.
* Leverage go's new fs.FS interface for collecting files.

## API Stability

This package has not yet released to 1.x yet, so APIs are subject to change for
a bit longer.

## Extensions

These are just the extensions the goschtalt team maintains.  Others may be available
and it's fairly easy to write your own.  Extensions have their own go.mod files
that independently track dependencies to keep dependencies only based on what
you need, not what could be used.

### Configuration Decoders

The decoders convert a file format into a useful object tree.  The meta.Object has
many convenience functions that make adding decoders pretty simple.  Generally,
the hardest part is determining where you are processing in the original file.

| Status | GoDoc | Extension | Description |
|--------|-------|-----------|-------------|
| [![Go Report Card](https://goreportcard.com/badge/github.com/goschtalt/goschtalt/extensions/decoders/env)](https://goreportcard.com/report/github.com/goschtalt/goschtalt/extensions/decoders/env) | [![GoDoc](https://pkg.go.dev/badge/github.com/goschtalt/goschtalt/extensions/decoders/env)](https://pkg.go.dev/github.com/goschtalt/goschtalt/extensions/decoders/env) | `decoders/env` | An environment variable based configuration decoder. |
| [![Go Report Card](https://goreportcard.com/badge/github.com/goschtalt/goschtalt/extensions/decoders/json)](https://goreportcard.com/report/github.com/goschtalt/goschtalt/extensions/decoders/json) | [![GoDoc](https://pkg.go.dev/badge/github.com/goschtalt/goschtalt/extensions/decoders/json)](https://pkg.go.dev/github.com/goschtalt/goschtalt/extensions/decoders/json) | `decoders/json` | A JSON configuration decoder. |
| [![Go Report Card](https://goreportcard.com/badge/github.com/goschtalt/goschtalt/extensions/decoders/properties)](https://goreportcard.com/report/github.com/goschtalt/goschtalt/extensions/decoders/properties) | [![GoDoc](https://pkg.go.dev/badge/github.com/goschtalt/goschtalt/extensions/decoders/properties)](https://pkg.go.dev/github.com/goschtalt/goschtalt/extensions/decoders/properties) | `decoders/properties` | A properties configuration decoder. |
| [![Go Report Card](https://goreportcard.com/badge/github.com/goschtalt/goschtalt/extensions/decoders/yaml)](https://goreportcard.com/report/github.com/goschtalt/goschtalt/extensions/decoders/yaml) | [![GoDoc](https://pkg.go.dev/badge/github.com/goschtalt/goschtalt/extensions/decoders/yaml)](https://pkg.go.dev/github.com/goschtalt/goschtalt/extensions/decoders/yaml) | `decoders/yaml` | A YAML/YML configuration decoder |


### Configuration Encoders

The encoders are used to output configuration into a file format.  Ideally you want
a format that accepts comments so it's easier see where the configurations originated
from.

| Status | GoDoc | Extension | Description |
|--------|-------|-----------|-------------|
| [![Go Report Card](https://goreportcard.com/badge/github.com/goschtalt/goschtalt/extensions/encoders/yaml)](https://goreportcard.com/report/github.com/goschtalt/goschtalt/extensions/encoders/yaml) | [![GoDoc](https://pkg.go.dev/badge/github.com/goschtalt/goschtalt/extensions/encoders/yaml)](https://pkg.go.dev/github.com/goschtalt/goschtalt/extensions/encoders/yaml) | `encoders/yaml` | A YAML/YML configuration encoder. |


## Dependencies

There are only two production dependencies in the core goschtalt code beyond the
go standard library.  The rest are testing dependencies.

Production dependencies:

* github.com/mitchellh/hashstructure
* github.com/mitchellh/mapstructure

## Examples

Coming soon.
