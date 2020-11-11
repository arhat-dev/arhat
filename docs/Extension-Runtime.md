# Runtime

Kubernetes defines its runtime API as `CRI` (container runtime interface)

## Design

## Configuration

__NOTE:__ this configuration section is ignored by `arhat`, which has no container runtime support

```yaml
runtime:
  # default: true
  #
  # if set to false, will create a empty no-op runtime
  enabled: true
```
