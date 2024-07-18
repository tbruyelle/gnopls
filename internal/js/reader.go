package js

import (
	"context"
	"io"
)

var _ io.Reader = (*ChannelReader)(nil)

// ChannelReader is a reader that reads from a channel of bytes.
type ChannelReader struct {
	buff     []byte
	source   <-chan []byte
	cancelFn context.CancelFunc
}

// NewChannelReader creates a new ChannelReader from a channel.
//
// Second argument is optional function that will be called when `Close` method is called.
func NewChannelReader(source <-chan []byte, cancelFn context.CancelFunc) *ChannelReader {
	return &ChannelReader{
		source: source,
		cancelFn: cancelFn,
	}
}

func (listener *ChannelReader) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}

	if len(listener.buff) == 0 {
		if err := listener.fetchMore(); err != nil {
			return 0, err
		}
	}

	readCount := min(len(listener.buff), len(b))

	copy(b, listener.buff[:readCount])
	listener.buff = listener.buff[readCount:]
	return readCount, nil
}

func (listener *ChannelReader) Close() error {
	if listener.cancelFn != nil {
		listener.cancelFn()
	}

	return nil
}

func (listener *ChannelReader) fetchMore() error {
	select {
	case message, ok := <-listener.source:
		if !ok {
			return io.EOF
		}

		listener.buff = message
		return nil
	}
}
