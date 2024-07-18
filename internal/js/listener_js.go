package js

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"syscall/js"
)

const (
	// jsFieldCallbackID is field name of js.Func struct that contains callback ID of Go func exposed to JS.
	jsFieldCallbackID = "id"

	// msgBufferSize is size of a buffer of incoming messages from LSP client.
	msgBufferSize = 10
)

// registerRequestListener creates and registers LSP request listener at the host.
// This method should only be called once at start of a server.
//
// Context parameter is used to automatically dispose underlying channel and callback.
//
// Returns a reader to consume incoming JSON-RPC messages from host.
// Closing a returned reader acts the same as cancelling an input context.
func registerRequestListener(ctx context.Context) (io.ReadCloser, error) {
	chanCtx, cancelFn := context.WithCancel(ctx)

	inputEvents := make(chan []byte, msgBufferSize)
	callback := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) == 0 {
			// this is the only way to throw a JS exception.
			panic("missing argument")
		}
		message := args[0].String()
		inputEvents <- []byte(message)
		return nil
	})

	// As it's impossible to pass JS objects to import funcs.
	// Obtain the function callback ID pass it to import.
	callbackID, err := getFuncCallbackID(callback)
	if err != nil {
		cancelFn()
		callback.Release()
		close(inputEvents)
		return nil, fmt.Errorf("failed to prepare callback: %w", err)
	}

	registerCallback(callbackID)

	go func() {
		<-chanCtx.Done()
		callback.Release()
		close(inputEvents)
	}()

	reader := NewChannelReader(inputEvents, cancelFn)
	return reader, nil
}

// getFuncCallbackID obtains callback ID from wrapped js.Func value.
//
// Internally, Go stores each js.Func handler inside a special lookup table.
// When host calls and wrapped function (js.Func), it resumes Go program and passes a callback ID.
// Go matches the callback ID with the corresponding js.Func handler in the lookup table and calls it.
//
// This flow didn't change since first Go with WASM support initial release.
//
// See: handleEvent in /syscall/js/js.go
func getFuncCallbackID(fn js.Func) (id uint32, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
		}
	}()

	// Obtain the callback id from the function.
	// Use reflection to capture possible struct layout changes.
	field := reflect.ValueOf(fn).FieldByName(jsFieldCallbackID)
	if !field.IsValid() {
		return 0, fmt.Errorf("cannot find field %q in %T", jsFieldCallbackID, fn)
	}

	if field.Type().Kind() != reflect.Uint32 {
		return 0, fmt.Errorf(
			"unexpected %T.%s field type: %s (want: %T)",
			fn, jsFieldCallbackID, field.Type(), id,
		)
	}

	id = uint32(field.Uint())
	if id == 0 {
		return 0, fmt.Errorf("empty callback ID in %T.%s", fn, jsFieldCallbackID)
	}

	return id, nil
}
