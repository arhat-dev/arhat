# Build

__TL;DR:__ have a look at the `build-on-{linux,darwin,windows}` jobs in the [build workflow](./.github/workflows/build.yaml)

## Binary Targets (`make` targets)

For native build: `make arhat`

Target format `arhat.{os}.{arch}`

To find all binary targets, run:

```bash
make -qp \
  | awk -F':' '/^[a-zA-Z0-9][^$#\/\t=]*:([^=]|$)/ {split($1,A,/ /);for(i in A)print A[i]}' \
  | sort -u \
  | grep ^arhat
```

## Build Tags

Set build tags using environment variable `make arhat TAGS='<space separated build tags>'`

### Functionality Build Tags

- `nosysinfo`
  - Disable system resource (memory, disk, cpu, etc.) report
- `noexectry` (save ~3MB space)
  - Disable all internal handling of command execution, to disable specific command handling, use following build tags:
    - `noexectry_tar` (save ~3MB space)
      - Disable internal handling of `tar`, this command handling is useful for those busybox based rootfs or windows with no standard GNU `tar` installed
      - NOTE: `kubectl cp` uses `tar` command execution to copy file
    - `noexectry_test`
      - Disable internal handling of `test`, this command handling is useful for windows without `test` command support for `kubectl cp`
      - NOTE: `kubectl cp` will invoke `test` command to check whether destination path is a directory
- `nometrics` (save ~4MB space)
  - Disable metrics collection, no node metrics or peripheral metrics will be collected
- Build tags from [prometheus/node_exporter](https://github.com/prometheus/node_exporter) for collectors
  - format: `no<collector-name>`
- `noextension`
  - Disable all extension support, to disale specific extension support, use following build tags:
    - `noextension_peripheral`
      - Disable peripheral extension support
    - `noextension_runtime`
      - Disable runtime extension support
- Build tags from [ext.arhat.dev/runtimeutil/storageutil](https://github.com/arhat-ext/runtimeutil-go/blob/master/storageutil) for storage drivers
  - format: `nostorage_<driver-name>`
  - values: `nostorage_general`, `nostorage_sshfs`
- Build tags from [arhat.dev/pkg/nethelper](https://github.com/arhat-dev/go-pkg/blob/master/nethelper) for network protocols
  - This controls network protocols supported by the embedded extension hub and port-forward
  - format: `nonethelper_<pkg-name>`
  - values: `nonethelper_stdnet` (tcp/udp/unix), `nonethelper_piondtls` (udp with dtls), `nonethelper_pipenet` (fifo pairs)

### Connectivity Method Build Tags

- `noclient_grpc` (save ~3MB space)
  - Disable `gRPC` connectivity support
- `noclient_mqtt` (save ~1MB space)
  - Disable `MQTT` connectivity support
- `noclient_coap` (save ~1MB space)
  - Disable `CoAP` connectivity support
