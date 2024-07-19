interface ConnectParams {
  /**
   * MessagePort to use by server to communicate with LSP client.
   */
  port: MessagePort
}


/**
 * LSPWorker defines a Comlink worker interface.
 */
export interface LSPWorker {
  connect: (args: ConnectParams) => Promise<void>
}