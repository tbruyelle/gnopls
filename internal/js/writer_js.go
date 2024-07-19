package js

import "unsafe"

// messageWriter passes JSON-RPC messages to WASM host.
var messageWriter = wasmWriter{
	writeFunc: writeMessage,
	closeFunc: closeWriter,
}

type wasmWriter struct {
	writeFunc func(p unsafe.Pointer)
	closeFunc func()
}

func (w wasmWriter) Write(p []byte) (n int, err error) {
	w.writeFunc(unsafe.Pointer(&p))
	return len(p), nil
}

func (w wasmWriter) Close() error {
	if w.closeFunc != nil {
		w.closeFunc()
	}

	return nil
}
