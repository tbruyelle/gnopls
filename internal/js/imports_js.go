package js

import "unsafe"

//go:wasmimport lsp writeMessage
func writeMessage(p unsafe.Pointer)

//go:wasmimport lsp closeWriter
func closeWriter()

//go:wasmimport lsp registerCallback
func registerCallback(callbackID uint32)
