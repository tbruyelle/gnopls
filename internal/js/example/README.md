# Gnopls Embedding Example

This directory contains a bare-minimum code to integrate Gnopls as a WebAssembly module
into browser or Node.js environment.

This example omits such nuances as editor integration or file system support and focuses just on basics.

## Prerequisites

* Copy `wasm_exec.js` file using following command:
    * `cp $(go env GOROOT)/misc/wasm/wasm_exec.js .`
* Build gnopls as a WebAssembly file for JavaScript environment:
    * `GOOS=js GOARCH=wasm make build`
* Modify paths to `wasm_exec.js` and WASM file in [worker.ts](./worker.ts) file.