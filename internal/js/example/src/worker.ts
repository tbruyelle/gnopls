/**
 * This file contains a web worker used to host and configure LSP server.
 */

import * as Comlink from 'comlink'

import { LSPWorker } from './types';
import { configureGoInstance } from './internal/setup';

declare const self: DedicatedWorkerGlobalScope

// Don't forget to copy wasm_exec.js from '$GOROOT/misc/wasm'
importScripts('/wasm_exec.js')

// Don't forget to build gnopls
const GNOPLS_URL = '/gnopls.wasm'

const worker: LSPWorker = {
  connect: async ({ port }) => {
    // For browsers: don't forget to configure filesystem before using gnopls!
    // Override globalThis.fs with your own implementation in order to provide source files for a server.
    const go = new Go()

    // Feel free to pass custom environment variables and cmdline args
    // go.env.FOO = 'bar'
    // go.argv = ['gopls']

    configureGoInstance(go, port)

    // Fetch the worker and instantiate a wasm instance
    const { instance } = await fetch(GNOPLS_URL)
      .then((rsp) => WebAssembly.instantiateStreaming(rsp, go.importObject))

    // Start the server in background
    go.run(instance)
      .then((code) => console.log('gnopls: server exited with code ', code))
      .catch((err) => console.error('gnopls: cannot start server: ', err))
  },
}

Comlink.expose(worker);
