//go:build !js

package env

import (
	"context"
	"net"
	"os"

	"go.lsp.dev/pkg/fakenet"
)

func GetConnection(_ context.Context) (net.Conn, error) {
	return fakenet.NewConn("stdio", os.Stdin, os.Stdout), nil
}
