package defs

import "io"
import "github.com/gorilla/websocket"

const (
	// TextWriter asks the nextwriter for a text based writer
	TextWriter = websocket.TextMessage
)

// Streamer defines an interface that allows consumers to open a writer, reader and close the connection
type Streamer interface {
	NextWriter(int) (io.WriteCloser, error)
	Close() error
	NextReader() (int, io.Reader, error)
}
