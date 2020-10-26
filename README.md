# arhat `阿罗汉`

[![CI](https://github.com/arhat-dev/arhat/workflows/CI/badge.svg)](https://github.com/arhat-dev/arhat/actions?query=workflow%3ACI)
[![Build](https://github.com/arhat-dev/arhat/workflows/Build/badge.svg)](https://github.com/arhat-dev/arhat/actions?query=workflow%3ABuild)
[![PkgGoDev](https://pkg.go.dev/badge/arhat.dev/arhat)](https://pkg.go.dev/arhat.dev/arhat)
[![GoReportCard](https://goreportcard.com/badge/arhat.dev/arhat)](https://goreportcard.com/report/arhat.dev/arhat)
[![codecov](https://codecov.io/gh/arhat-dev/arhat/branch/master/graph/badge.svg)](https://codecov.io/gh/arhat-dev/arhat)

The reference `EdgeDevice` agent for `aranya`

## Features

- Detailed and Customizable Node Infromation Report
  - Customize Kubernetes `Node` annotations and labels in conjunction with `aranya`
- Flexible connectivity
  - `MQTT 3.1.1` (including `aws-iot-core`, `azure-iot-hub`, `gcp-iot-core`)
  - `CoAP` (including `tcp`, `tcp/tls`, `udp`, `udp/dtls`)
  - `gRPC`
- Extensible plugin system
  - Create your own peripheral controller and integrate into Kubernetes API with ease (e.g. use `kubectl` to operate peripheral)
- Unified Metrics Collection
  - Collect all kinds of metrics from `aranya` and Kubernetes API, including metrics of peripherals

## Design

- [Connectivity](./docs/Connectivity.md)
- [Runtime](./docs/Runtime.md)
- [Networking](./docs/Networking.md)
- [Storage](./docs/Storage.md)
- [Extension](./docs/Extension.md)

## Configuration

see [docs/Configuration](./docs/Configuration.md)

## Build

__TL;DR:__ have a look at the `build-on-{linux,darwin,windows}` jobs in [build workflow](./.github/workflows/build.yaml)

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
