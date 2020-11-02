# Extension

Edge device usually comes with special peripherals like sensors for data collecting, we would like to address this to provide a unified way to integrate existing peripherals into Kubernetes

__TL;DR:__ To get started with your own extension plugin, we recommend you having a look at [this extension template project (golang)](https://github.com/arhat-dev/template-arhat-ext-go)

## Purpose

- Interact with your peripherals via Kubernetes API (e.g. kubectl exec)
- Collect prometheus metrics from your peripherals with simple string key value args

## Design

There are two major parts contributing to the extension system:

- `extension server` (actuall a http server, embedded in `arhat`)
- `extension plugin` (actuall a http client)

The interaction between the `arhat` and `extension plugin`:

- `extension server` listen on one port (`tcp`/`unix`) as configured
- `extension plugin` connect to the address that `extension server` listening
  - once connected, `extension plugin` will post a http request, but different from the normal http request, this post request will never end unless interrupped by user or error. To put this simple: send a infinite post request to `extension server`
- `extension server` knows how to handle this kind of post request, and maintains it until network error happened
- `extension plugin` will register itself with a unique name (locally to `extension server`) in the initial post request
- `extension server` will interact with registered `extension plugin` when necessary (e.g. received certain command from message queue)

## Extension Endpoints

Extension endpoints are just http paths

### `/peripherals`: Interact with physical world

- This extension is designed to support operations and metrics collections for all kinds of physical peripherals, sensors, routers...
- More peripheral extension apps are comming soon

## Configuration

The extension configuration defines the functionality of [extension api (arhat-proto)](https://github.com/arhat-dev/arhat-proto) server

```yaml
extension:
  # enable extension service or not
  enabled: true
  
  endpoints:
    # listen address of the endpoint
    # supported protocols: unix, tcp/tcp4/tcp6, udp/udp4/udp6
    #
    # if you want to serve multiple protocols/addresses,
    # just add multiple endpoints with different listen address
  - listen: unix:///var/run/arhat.sock
    # tls configuration for the endpoint (server tls)
    tls:
      # enable tls or not
      enabled: false

    # keepalive interval, defaults to one minute
    keepaliveInterval: 1m

    # how long should we wait for message response, defaults to one minute
    messageTimeout: 1m

  - listen: udp://localhost:65432
    # dtls is used for udp with tls enabled
    tls:
      # enable tls or not
      enabled: true

      # require client certificate or not
      verifyClientCert: false

      # CA cert file (PEM/ASN.1 format)
      caCert: /path/to/ca.crt
      # You can specify base64 encoded ca cert data directly as an alternative to caCert
      #caCertData: "<base64-encoded-ca-cert>"

      # client cert file (PEM format)
      #
      # for variant `gcp-iot-core`, this field MUST be empty
      cert: /path/to/client.crt
      # You can specify base64 encoded cert data directly as an alternative to cert
      #certData: "<base64-encoded-tls-cert>"

      # client private key file (PEM format)
      key:  /path/to/client.key
      # You can specify base64 encoded tls key directly as an alternative to `key`
      #keyData: "<base64-encoded-tls-key>"

      # tls server name override
      serverName: foo.example.com

      # ONLY intended for DEBUG use
      #
      # if set, will record the random key used in the tls connection to this file
      # and the file can be used for applications like wireshark to decrypt tls connection
      keyLogFile: /path/to/tmp/tls/log

      # set cipher suites expected to ues when establishing tls connection
      #
      # please refer to Appendix.A section in Connectivity.md for full
      # list of supported cipher suites
      cipherSuites: []
      # - TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256

  # peripheral extension config
  peripherals:
    # cache unhandled metrics for at most this time
    metricsCacheTimeout: 1h
```
