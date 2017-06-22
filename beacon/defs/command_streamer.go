package defs

import "io"

// CommandStreamer defines an interface that allows consumers to open a writer, reader and close the connection
type CommandStreamer interface {
	NextWriter(int) (io.WriteCloser, error)
	Close() error
	NextReader() (int, io.Reader, error)
}
