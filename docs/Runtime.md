# Runtime

## Configuration

__NOTE:__ this configuration section is ignored by `arhat-none`, which has no container runtime support

```yaml
runtime:
  # enable bundled container runtime
  #
  # default: true
  #
  # if set to false, will create a empty no-op runtime
  enabled: true

  # the directory to store pod data and container specific files
  #
  # generally it will include:
  #   - volume data from configmap/secret
  #   - directory created by emptyDir volume
  #   - container resolv.conf
  dataDir: /var/lib/arhat

  # managementNamespace is the container runtime's namespace
  #
  # used in:
  #   arhat-libpod
  # ignored in:
  #   arhat-docker
  managementNamespace: container.arhat.dev

  # pauseImage is the image used to create the first container in pod
  # to claim all the linux namespaces required by the pod
  pauseImage: k8s.gcr.io/pause:3.1

  # pauseCommand is the command to run pause image
  pauseCommand: /pause

  # service endpoints of the container runtime
  endpoints:

    # image endpoint (image management)
    image:
      # the url to connect the image endpoint
      #
      # used in:
      #   - arhat-docker
      # ignored in:
      #   - arhat-libpod
      endpoint: unix:///var/run/docker.sock

      # timeout when dial image endpoint
      #
      # used in:
      #   - arhat-docker
      # ignored in:
      #   - arhat-libpod
      dialTimeout: 10s

      # timeout when working on image related job (image pull and lookup)
      actionTimeout: 2m

    # container endpoint
    container:
      # the url to connect the container endpoint
      #
      # used in:
      #   - arhat-docker
      # ignored in:
      #   - arhat-libpod
      endpoint: unix:///var/run/docker.sock

      # timeout when dial runtime endpoint
      #
      # used in:
      #   - arhat-docker
      # ignored in:
      #   - arhat-libpod
      dialTimeout: 10s

      # timeout when working on container related job
      actionTimeout: 2m
```
