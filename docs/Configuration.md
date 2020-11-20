# `arhat` Config

- [Overview](#overview)
  - [Section `arhat`](#section-arhat)
  - [Section `connectivity`](#section-connectivity)
  - [Section `storage`](#section-storage)
  - [Section `extension`](#section-extension)

## Overview

The configuration of `arhat` is defined in a `yaml` or `json` file

The `arhat` configuration file contains four major sections (`arhat`, `storage`, `connectivity` and `extension`):

```yaml
# arhat section controls application behavior
arhat:
  # ...
storage:
  # ...
connectivity:
  # ...
extension:
  # ...
```

Each section is used to define a specific aspect of `arhat`'s behavior

__NOTE:__ You can include environment variables (`$FOO`, `$(FOO)` or `${FOO}`) in the config file, they will be expanded according to system environment variables when `arhat` loading configuration, if the environment variable was not defined, `arhat` will keep it as is (say, `$FOO` will be `$FOO`, not empty string or `${FOO}`).

### Section `arhat`

This section defines `arhat` application behavior

```yaml
arhat:
  # log options, you can designate mutiple destination as you wish
  log:
    # log level
    #
    # can be one of the following:
    #   - verbose
    #   - debug
    #   - info
    #   - error
    #   - silent  (no log)
  - level: debug
    # log output format
    #
    # value can be one of the following:
    #   - console
    #   - json
    format: console
    # log output destination file
    #
    # value can be one of the following:
    #   - stdout
    #   - stderr
    #   - {FILE_PATH}
    file: stderr
    # whether expose this log file for `kubectl logs`
    #
    # if true, `kubectl logs` to the according virtual pod will
    # use this file as log source
    #
    # if multiple log config section has `kubeLog: true`, the
    # last one in the list will be used
    kubeLog: false
  - level: debug
    format: console
    kubeLog: true
    file: /var/log/arhat.log

  # host operations
  # for security reason, all of them are set to `false` by defualt
  host:
    # rootfs into other dir after agent is running, useful for host
    # management when deployed as container
    #
    # NOTE: this will make `kubectl logs` to the virtual pod not working
    #       unless there is a symlink to kubeLog file in the new root
    rootfs: ""

    # set user id and group id
    uid: 1000
    gid: 1000

    # allow `kubectl exec/cp` to device host
    allowExec: true
    # allow `kubectl attach` to device host
    allowAttach: true
    # allow `kubectl port-forward` to device host
    allowPortForward: true
    # allow `kubectl logs` to view arhat log file exposed with `kubeLog: true`
    allowLog: true

  # kubernetes node operation
  node:
    # set custom machine id, if not set, will report standard machine id as kubelet will do
    # if not set, or value is empty, will pick up standard platform specific machine id automatically
    machineIDFrom:
      # Execute a command and use output as machine id
      #exec: []
      # Read file content as machine id
      #file: ""
      # Set machine id explicitly
      text: "foo"

    # extInfo a list of key value pairs to set extra node related values
    extInfo:
      # resolved value must be a string, no matter what type it is
    - valueFrom:
        # Execute a command and use output as value
        #exec: []
        # Read file content as value
        #file: ""
        # Set value explicitly
        text: "1"

      # operator available: [=, +=, -=]
      operator: +=
      # valueType available: [string, int, float]
      valueType: int
      # applyTo which node object field
      # value available: [metadata.annotations[''], metadata.labels['']]
      applyTo: metadata.annotations['example.com/key']

  # pprof module, disabled when built with `noconfhelper_pprof`
  pprof:
    # enable pprof
    enabled: false
    # pprof listen address
    listen: localhost:8080
    # http base path
    httpPath: /debug/pprof
    # parameter for `runtime.SetCPUProfileRate(int)`
    cpuProfileFrequencyHz: 100
    # parameter for `runtime.SetMutexProfileFraction(int)`
    mutexProfileFraction: 100
    # parameter for `runtime.SetBlockProfileRate(int)`
    blockProfileFraction: 1
```

### Section `connectivity`

This section defines how to connect to the server/broker and finally communicate with `aranya`

see [Connectivity Configuration](./Connectivity.md#configuration)

### Section `storage`

This section defines how to mount remote volumes

see [Storage Configuration](./Storage.md#configuration)

### Section `extension`

This section defines behavior of the embedded extension server

see [Extension Configuration](./Extension.md#configuration)
