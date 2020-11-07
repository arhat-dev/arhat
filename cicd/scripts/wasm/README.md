# wasm

Run `arhat` in browser

## Files

- `config.yaml` - Sample config with mqtt websocket connection
- `index.html` - Sample html page to load `arhat.js.wasm`
- `wasm_exec.js` - Sample wasm runtime setup script
- `arhat.js.wasm` - The wasm binary

## Example Setup

1. Build the wasm binary with `make arhat.js.wasm`, you can find the binary in `{PROJECT_ROOT}/build`, copy it to this directory:

    ```bash
    # in PROJECT_ROOT
    make arhat.js.wasm
    mv build/arhat.js.wasm cicd/scripts/wasm/arhat.js.wasm
    ```

2. Create the `wasm_exec.js` according to the version of the `go` toolchain you used to build `arhat.js.wasm` (run `go version`, only MAJOR and MINOR version number used):

    ```bash
    # in PROJECT_ROOT
    cp cicd/scripts/wasm/wasm_exec_go{MAJOR}.{MINOR}.js cicd/scripts/wasm/wasm_exec.js
    ```

3. Update `config.yaml` according to your own `aranya` and `EdgeDevice` setup (please refer to [`docs/Configuration`](../../../docs/Configuration.md) for all configuration options)

4. Run a web server to serve all these [files](#files):

    ```bash
    # in PROJECT_ROOT
    docker run -it -p 8080:8080 -v $(pwd)/cicd/scripts/wasm:/app bitnami/nginx:latest
    ```

5. Open your browser and navigate to [`http://localhost:8080/`](http://localhost:8080/) and you will find the log output on the web page
