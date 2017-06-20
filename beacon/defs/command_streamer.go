package defs

import "io"

type CommandStreamer interface {
	NextWriter(int) (io.WriteCloser, error)
	Close() error
	NextReader() (int, io.Reader, error)
}
