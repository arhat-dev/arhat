# libserial

[![Build Status](https://travis-ci.com/goiiot/libserial.svg)](https://travis-ci.com/goiiot/libserial) [![GoDoc](https://godoc.org/github.com/goiiot/libserial?status.svg)](https://godoc.org/github.com/goiiot/libserial) [![GoReportCard](https://goreportcard.com/badge/goiiot/libserial)](https://goreportcard.com/report/github.com/goiiot/libserial) [![codecov](https://codecov.io/gh/goiiot/libserial/branch/master/graph/badge.svg)](https://codecov.io/gh/goiiot/libserial)

Serial library for golang (no cgo)

## Prerequisite

- Go 1.9+ (for `type alias`)
- Git (required by Go)

## Supported Platform

- darwin (tested on amd64, thanks to [`travis`](https://travis-ci.com/goiiot/libserial))
  - arm64 amd64 arm 386
- linux (tested on amd64, thanks to [`travis`](https://travis-ci.com/goiiot/libserial))
  - all go supported arch
- windows (not tested, see [issue #1](https://github.com/goiiot/libserial/issues/1))
  - all go supported arch
- freebsd (not tested)
  - all go supported arch
- netbsd (not tested)
  - all go supported arch
- openbsd (not tested)
  - all go supported arch

## Usage

**TL;DR**: you can find a demo cli app in [cmd/libserial/main.go](./cmd/libserial/main.go)

0.Get this package with `go get` or `git clone`

```bash
go get -u github.com/goiiot/libserial
# git clone https://github.com/goiiot/libserial
```

1.Import this package

```go
import (
    // ...
    "github.com/goiiot/libserial"
)
```

2.Open serial connection and check error

```go
// open serial port with default settings (9600 8N1)
conn, err := libserial.Open("/dev/serial0")

if err != nil {
    panic("hmm, how cloud it fail")
}
```

**Note**: You can add options when opening serial port, see [godoc - Option](https://godoc.org/github.com/goiiot/libserial#Option)

3.Read/Write data from serial connection

```go
buf := make([]byte, 64)

_, err := conn.Read(buf[:])
if err != nil { }

_, err := conn.Write([]byte("{data}"))
if err != nil { }
```

## Command line demo

You can download and install `libserial` to your `$GOPATH/bin` for quick demo test (`GOPATH` required)

```bash
go get -u github.com/goiiot/libserial/cmd/libserial
```

## References

- [termios(3) - Linux man page](https://linux.die.net/man/3/termios)
- [tty(4) - FreeBSD Manual Page](https://www.freebsd.org/cgi/man.cgi?query=tty&sektion=4)

## LICENSE

[![GitHub license](https://img.shields.io/github/license/goiiot/libserial.svg)](https://github.com/goiiot/libserial/blob/master/LICENSE.txt)

```text
Copyright Go-IIoT (https://github.com/goiiot)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
