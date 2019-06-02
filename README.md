# Overview
[![GoDoc](https://godoc.org/kreklow.us/go/t38c?status.svg)](https://godoc.org/kreklow.us/go/t38c) ![GitHub](https://img.shields.io/github/license/cjkreklow/t38c.svg) ![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/cjkreklow/t38c.svg) [![Build Status](https://www.travis-ci.org/cjkreklow/t38c.svg?branch=master)](https://www.travis-ci.org/cjkreklow/t38c) [![codecov](https://codecov.io/gh/cjkreklow/t38c/branch/master/graph/badge.svg)](https://codecov.io/gh/cjkreklow/t38c) [![Go Report Card](https://goreportcard.com/badge/kreklow.us/go/t38c)](https://goreportcard.com/report/kreklow.us/go/t38c)

`t38c` is a Go client library for the Tile38 geospatial database.

# Usage

`import kreklow.us/go/t38c`

Use `Connect()` to instantiate a `Database` object. A limited set of commands are currently available. Commands that retrieve data return a `Response` object.

Functions other than `Close()` accept arguments in the same form as the Tile38 CLI. See [Tile38 Commands](https://tile38.com/commands/) for further information.

# Links
 * [Tile38 Web Site](https://tile38.com/)

# About
`t38c` is maintained by Collin Kreklow. The source code is licensed under the terms of the MIT license, see `LICENSE.txt` for further information.
