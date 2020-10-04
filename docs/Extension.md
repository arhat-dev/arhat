# Extension

## Design

There are two major parts contributing to the extension system:

- `arhat` (a http server)
- `extension application` (a http client)

The interaction between the `arhat` and `extension application`:

- `arhat` listen one local address (`tcp`/`unixsock`) as configured
- `extension application`s connect to the address that `arhat` listening
- once connected to the address, `extension application` will post a http request, but different from the normal http request, this post request will never end unless interrupped by user or error. To put it simple: send a infinite post request to `arhat`
- `arhat` knows how to handle this kind of post request, and maintains it until network error happened
- `extension application` will register itself with a unique name (locally to arhat)
- `arhat` will interact with registered `extension application`s when necessary (e.g. received certain command from `aranya`)

## Supported Extensions

- Devices (http path: `/ext/devices`)
  - This extension is designed to support operations and metrics collections for all kinds of physical devices, sensors, routers...
  - To get started with your own customized device extension app, we recommend you having a look at [this device template](https://github.com/arhat-dev/template-arhat-device)
  - More device extension apps are comming soon
