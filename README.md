# Overview
[![PkgGoDev](https://pkg.go.dev/badge/kreklow.us/go/t38c)](https://pkg.go.dev/kreklow.us/go/t38c)
![License](https://img.shields.io/github/license/cjkreklow/t38c)
![Version](https://img.shields.io/github/v/tag/cjkreklow/t38c)
![Status](https://github.com/cjkreklow/t38c/actions/workflows/push.yml/badge.svg?branch=main)
[![Codecov](https://codecov.io/gh/cjkreklow/t38c/branch/main/graph/badge.svg)](https://codecov.io/gh/cjkreklow/t38c)

`t38c` is a Go client library for the Tile38 geospatial database.

# Usage

`import kreklow.us/go/t38c`

Use `Connect()` to instantiate a `Database` object. A limited set of
commands are currently available. Commands that retrieve data return a
`Response` object.

Functions other than `Close()` accept arguments in the same form as the
Tile38 CLI. See [Tile38 Commands](https://tile38.com/commands/) for further
information.

# Links
 * [Tile38 Web Site](https://tile38.com/)

# About
`t38c` is maintained by Collin Kreklow. The source code is licensed under
the terms of the MIT license, see `LICENSE.txt` for further information.
