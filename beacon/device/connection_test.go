package device

import "io"
import "io/ioutil"
import "bytes"
import "testing"
import "github.com/satori/go.uuid"

type testBuffer struct {
	*bytes.Buffer
}

func (b *testBuffer) Close() error {
	return nil
}

type testStreamer struct {
	b *testBuffer
}

func (t *testStreamer) NextWriter(int) (io.WriteCloser, error) {
	return t.b, nil
}

func (t *testStreamer) NextReader() (int, io.Reader, error) {
	return 0, t.b, nil
}

func (t *testStreamer) Close() error {
	return nil
}

func Test_Device_Connection(suite *testing.T) {
	id := uuid.NewV4()

	connection := func(b *testBuffer) *Connection {
		return &Connection{&testStreamer{b}, id}
	}

	suite.Run("returns the buffer from NextReader", func(test *testing.T) {
		b := bytes.NewBuffer([]byte("hello world"))
		c := connection(&testBuffer{b})
		_, r, _ := c.NextReader()
		s, _ := ioutil.ReadAll(r)

		if string(s) != "hello world" {
			test.Fatalf("expected to read from buffer")
		}
	})

	suite.Run("returns the buffer from NextWriter", func(test *testing.T) {
		b := bytes.NewBuffer([]byte{})
		c := connection(&testBuffer{b})
		w, _ := c.NextWriter(0)
		io.Copy(w, bytes.NewBuffer([]byte("goodbye")))

		if string(b.Bytes()) != "goodbye" {
			test.Fatalf("expected goodbye but got: %s", b.Bytes())
		}
	})

	suite.Run("returns the uuid string from GetID", func(test *testing.T) {
		b := bytes.NewBuffer([]byte{})
		c := connection(&testBuffer{b})
		e, a := id.String(), c.GetID()

		if e != a {
			test.Fatalf("expected %s but got %s", e, a)
		}
	})
}
