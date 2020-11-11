# Build

__TL;DR:__ have a look at the `build-on-{linux,darwin,windows}` jobs in the [build workflow](./.github/workflows/build.yaml)

## Build Tags

Set build tags using environment variable `TAGS='<some tags>'`

- `noexectry` (save ~3MB space)
  - Disable `tar`, `test` command internal handling (which is useful for busybox rootfs without standard gnu `tar` implementation)
- `nometrics` (save ~4MB space)
  - Disable metrics collection, no node metrics and peripheral metrics will be collected
- `nogrpc` (save ~3MB space)
  - Disable `gRPC` connectivity support
- `nomqtt` (save ~1MB space)
  - Disable `MQTT` connectivity support
- `nocoap` (save ~1MB space)
  - Disable `CoAP` connectivity support
- `noext`
  - Disable extension support (no runtime/peripheral integration)

### Binary Targets

Target format `arhat.{os}.{arch}`

To find all binary targets, run:

```bash
make -qp \
  | awk -F':' '/^[a-zA-Z0-9][^$#\/\t=]*:([^=]|$)/ {split($1,A,/ /);for(i in A)print A[i]}' \
  | sort -u \
  | grep ^arhat
```
