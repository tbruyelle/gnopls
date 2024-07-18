/**
 * This is a simple demo file that shows how to configure a gopls Go instance
 * for communication between LSP client and server.
 *
 * Implementation details may vary depending on editor library.
 *
 * For Monaco, please check "@codingame/monaco-editor-treemended" package.
 * For CodeMirror - https://github.com/FurqanSoftware/codemirror-languageserver
 */

import { MonacoLanguageClient } from 'monaco-languageclient';
import { BrowserMessageReader, BrowserMessageWriter } from 'vscode-languageserver-protocol/browser'
import { CloseAction, ErrorAction } from 'vscode-languageclient'

import * as Comlink from 'comlink';

import { LSPWorker } from './types';

// See: worker.ts
const lspWorker = new Worker('worker.js')

// Create a pair of message ports dedicated only for LSP messages.
const { port1: clientPort, port2: serverPort } = new MessageChannel();

// Comlink provides a convenient way of calling worker functions.
// Feel free to use plain "postMessage" calls instead.
const proxy = Comlink.wrap<LSPWorker>(lspWorker);

// Ask worker to start gnopls and use passed port to json-rpc communication.
// MessagePort should be passed as a transferable object.
await proxy.connect(Comlink.transfer({ port: serverPort }, [ serverPort ]))

// Feel free to pick up any LSP client library as long as it supports way to specify custom message transports.
// This example uses LSP client used by Monaco editor and VSCode.
const reader = new BrowserMessageReader(clientPort)
const writer = new BrowserMessageWriter(clientPort)

const lspClient = new MonacoLanguageClient({
    name: 'gnopls-lsp-client',
    clientOptions: {
      // use a language id as a document selector
      documentSelector: [
        { language: 'go', scheme: 'file' },
        { language: 'gno', scheme: 'file' },
      ],
      // disable the default error handler
      errorHandler: {
        error: () => ({ action: ErrorAction.Continue }),
        closed: () => ({ action: CloseAction.DoNotRestart }),
      },
    },
    // create a language client connection to the server running in the web worker
    connectionProvider: {
      get: () => {
        return Promise.resolve({ reader, writer })
      },
    },
  })

// Don't forget to dispose all resources at the end.
await lspClient.dispose()
void proxy[Comlink.releaseProxy]()
lspWorker.terminate()
