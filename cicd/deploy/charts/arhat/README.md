# arhat

`arhat` is the reference `EdgeDevice` agent working with `aranya`

## Introduction

This is the official helm chart for [arhat](https://github.com/arhat-dev/arhat), you can deploy `arhat` to your clusters to manage kubernetes node host operations

## Prerequisites

- `helm` v3
- `Kubernetes` 1.15+

## Installing the Chart

```bash
helm install my-release arhat-dev/arhat
```

## Uninstalling the Chart

```bash
helm delete my-release
```

## Configuration

Please refer to the [`values.yaml`](https://github.com/arhat-dev/arhat/blob/master/cicd/deploy/charts/arhat/values.yaml)
