// Package js provides primitives to integrate language server with javascript environments such as browsers and Node.
package js

import (
	"context"
	"net"

	"go.lsp.dev/pkg/fakenet"
)

// DialHost registers LSP message listener in JavaScript host and returns connection to use by LSP server.
//
// This function should be called only once before starting LSP server.
func DialHost(ctx context.Context) (net.Conn, error) {
	reader, err := registerRequestListener(ctx)
	if err != nil {
		return nil, err
	}

	conn := fakenet.NewConn("js", reader, messageWriter)
	return conn, nil
}
