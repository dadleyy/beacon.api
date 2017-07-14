package defs

import "io"

// Signer defines an interface that simply returns a signed slice of bytes
type Signer interface {
	Sign(io.Writer, []byte) error
}
