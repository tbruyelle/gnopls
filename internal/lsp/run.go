package lsp

import (
	"context"
	"errors"
	"io"

	"github.com/gnolang/gnopls/internal/env"
	"go.lsp.dev/jsonrpc2"
)

func RunServer(ctx context.Context, e *env.Env) error {
	conn, err := env.GetConnection(ctx)
	if err != nil {
		return err
	}

	rpcConn := jsonrpc2.NewConn(jsonrpc2.NewStream(conn))
	handler := BuildServerHandler(rpcConn, e)
	stream := jsonrpc2.HandlerServer(handler)
	err = stream.ServeStream(ctx, rpcConn)
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}
