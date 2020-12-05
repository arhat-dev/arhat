# arhat `阿罗汉`

[![CI](https://github.com/arhat-dev/arhat/workflows/CI/badge.svg)](https://github.com/arhat-dev/arhat/actions?query=workflow%3ACI)
[![Build](https://github.com/arhat-dev/arhat/workflows/Build/badge.svg)](https://github.com/arhat-dev/arhat/actions?query=workflow%3ABuild)
[![PkgGoDev](https://pkg.go.dev/badge/arhat.dev/arhat)](https://pkg.go.dev/arhat.dev/arhat)
[![GoReportCard](https://goreportcard.com/badge/arhat.dev/arhat)](https://goreportcard.com/report/arhat.dev/arhat)
[![codecov](https://codecov.io/gh/arhat-dev/arhat/branch/master/graph/badge.svg)](https://codecov.io/gh/arhat-dev/arhat)

The reference `EdgeDevice` agent for `aranya`, serving as the connectivity hub for real world.

## Features

- Easy deployment for any platform, [even in browser](./cicd/scripts/wasm) (requires wasm support)
- No external dependencies (`socat`, `tar`) required for `kubectl cp/port-forward/exec/attach`
- Detailed and customizable node infromation report
  - Customize Kubernetes `Node` annotations and labels in conjunction with `aranya` in local configuration
- Flexible connectivity
  - `MQTT 3.1.1` (including `aws-iot-core`, `azure-iot-hub`, `gcp-iot-core`)
    - supports `tcp`, `tcp/tls`, `websocket`, `websocket/tls`
  - `CoAP`
    - supports `tcp`, `tcp/tls`, `udp`, `udp/dtls`
  - `gRPC`
- Extensible plugin system built with [`libext`](arhat.dev/libext)
  - Create your own peripheral controller and integrate all kinds of peripherals into Kubernetes API with ease (e.g. use `kubectl` to turn on/off lights)
- Unified metrics collection
  - Efficient built-in dynamic prometheus `node_exporter`/`windows-exporter` with no port exposed
  - Collect all kinds of metrics with `aranya` and Kubernetes API, including metrics from peripherals

## Project State

__TL;DR:__ Currently you can treate `araht` as a cloud native replacement of `sshd` with node metrics reporting support working with `aranya` through Kubernetes API

Currently State of functionalities:

- Stable:
  - port-forward
  - exec/attach
  - node metrics collecting
  - extended node info reporting
- Unstable (subject to design change):
  - peripheral extension
  - runtime extension
- Experimental (not fully supported):
  - networking with abbot
  - remote storage mount

## Design

__NOTE:__ This is not the final design, and docs may not reflect the actual implementation for now

- [Connectivity](./docs/Connectivity.md)
- [Networking](./docs/Networking.md)
- [Storage](./docs/Storage.md)
- [Extension](./docs/Extension.md) for runtime and peripheral

## Build

see [docs/Build](./docs/Build.md)

## Configuration

see [docs/Configuration](./docs/Configuration.md)

## Run

- As Kubernetes pod for host management, see [cicd/deploy/charts/arhat](./cicd/deploy/charts/arhat)
- As system daemon, see [cicd/scripts](./cicd/scripts)

## LICENSE

```text
Copyright 2020 The arhat.dev Authors.

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
