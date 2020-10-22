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

### `/etx/peripherals`: Interact with physical world

- This extension is designed to support operations and metrics collections for all kinds of physical peripherals, sensors, routers...
- More peripheral extension apps are comming soon
