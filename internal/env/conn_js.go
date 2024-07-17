package env

import (
	"context"
	"net"

	"github.com/gnolang/gnopls/internal/js"
)

func GetConnection(ctx context.Context) (net.Conn, error) {
	return js.DialHost(ctx)
}
