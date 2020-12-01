# Networking

Network of edge device usually involves vpn/tunnel management, and some of the edge device may not need such network at all, so we decided to offload networking ability to external component like `abbot` ([learn more about `abbot` project](https://github.com/arhat-dev/abbot))

__NOTE:__ `abbot` is a reference implementation to support remote network, you can implement your own solution by implementing the [`abbot-proto`](https://github.com/arhat-dev/abbot-proto) and documented behaviors in this doc, but for this document, we will call all such solution `abbot` for briefness.

## Design

General idea: `arhat` is just command executor to invoke external network component, the network options are encoded in protobuf bytes and passed to `abbot` as stdin data, so once got some prepared network options, `arhat` just invoke a command like `/path/to/abbot process` to get everything working.

By this way, `arhat` don't have to know `abbot-proto` at all, you can develop your protocol for custom use case.

### Host Network Desgin

- When your deploy `abbot` you can specify interfaces, which means
  - `abbot` can manage these interfaces when running to provide basic networking ability according to your requirements
- `aranya` lives in a Kubernetes namespace, thus
  - It knows all `EdgeDevice`s (the represent of a `arhat` instance) in this namespace, so `aranya` is able to discover their addresses by sending encoded `abbot-proto` to query host addresses
- `aranya` send encoded query (protobuf bytes) to `arhat`
  - `arhat` invoke `abbot` after encoding the query into base64 encoded string
- `abbot` processed the query
  - if successful: write base64 encoded protobuf bytes to stdout
  - if failed: write plain text to stderr
- `arhat` send error message or protobuf bytes received from stdio to `aranya`
- Finally, every edge device and `aranya` in the same namespace will be in a network mesh, then
  - your edge device can reach your Kubernetes cluster network
  - applications in your Kubernetes cluster can reach your edge devices

### Container Network Design

__NOTE:__ if `arhat` has no container runtime support (e.g. `arhat`), container network support is totally dropped

- `aranya` manages the `Node` resource, so it knows the pod CIDRs of the node, and these pod CIDRs are for connected `arhat`
- if `arhat` supports container runtime (`arhat-docker`, `arhat-libpod`), `aranya` will update container network config everytime `arhat` get connected to the `aranya`
  - `arhat` just invoke `abbot` to notify `abbot` with updated config
- once `aranya` decides to create pod in edge device, it will
  - issue image pull request
  - __issue infra container creation request__
    - if the pod is not using host network, this will create a sandbox network for the pod: `arhat` will execute `abbot` with base64 encoded protobuf bytes of provided network options
  - ... (other actions omitted)

## Configuration

```yaml
network:
  # enable abbot request processing
  enabled: false
  
  # command to invoke abbot for request processing
  #
  # with the reference abbot implementation (https://github.com/arhat-dev/abbot)
  # you need to specify the cmdline used to start abbot and its sub command `process`
  #
  # NOTE: you can also use `docker exec -i` (MUST provide stdin)
  abbotExec:
  - /path/to/abbot
  - -c
  - /path/to/abbot-config.yaml
  - process
```
