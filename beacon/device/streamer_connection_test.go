package device

import "io"
import "log"
import "fmt"
import "bytes"
import "testing"
import "github.com/franela/goblin"
import "github.com/satori/go.uuid"
import "github.com/dadleyy/beacon.api/beacon/defs"
import "github.com/dadleyy/beacon.api/beacon/logging"
import "github.com/dadleyy/beacon.api/beacon/interchange"

func newStreamerLogger() *logging.Logger {
	out := bytes.NewBuffer([]byte{})
	logger := log.New(out, "", 0)
	logger.SetFlags(0)
	return &logging.Logger{Logger: logger}
}

type testStreamerResponse struct {
	r io.Reader
	w io.WriteCloser
	e error
}

type testStreamer struct {
	responses []testStreamerResponse
}

func (t *testStreamer) Close() error {
	return nil
}

func (t *testStreamer) NextWriter(kind int) (io.WriteCloser, error) {
	if len(t.responses) == 0 {
		return nil, fmt.Errorf("no-reader")
	}

	r := t.responses[0]

	return r.w, r.e
}

func (t *testStreamer) NextReader() (int, io.Reader, error) {
	if len(t.responses) == 0 {
		return 0, nil, fmt.Errorf("no-reader")
	}

	r := t.responses[0]

	return 0, r.r, r.e
}

type testSigner struct {
	errors []error
}

func (t *testSigner) Sign(io.Writer, []byte) error {
	if len(t.errors) >= 1 {
		return t.errors[0]
	}
	return nil
}

type testStreamerConnectionScaffolding struct {
	connection StreamerConnection
	streamer   *testStreamer
	signer     *testSigner
}

type testWriteCloser struct {
	errors []error
}

func (t *testWriteCloser) Close() error {
	return nil
}

func (t *testWriteCloser) Write(b []byte) (int, error) {
	if len(t.errors) >= 1 {
		return 0, t.errors[0]
	}

	return 0, nil
}

func Test_StreamerConnection(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("Send", func() {
		var scaffold testStreamerConnectionScaffolding

		g.BeforeEach(func() {
			streamer := &testStreamer{}
			signer := &testSigner{}

			connection := StreamerConnection{
				LeveledLogger: newStreamerLogger(),
				Streamer:      streamer,
				Signer:        signer,
				id:            uuid.NewV4(),
			}

			scaffold = testStreamerConnectionScaffolding{
				connection: connection,
				streamer:   streamer,
				signer:     signer,
			}
		})

		g.It("returns an error when authentication is missing from message", func() {
			message := interchange.DeviceMessage{}
			e := scaffold.connection.Send(message)
			g.Assert(e.Error()).Equal(defs.ErrBadInterchangeAuthentication)
		})

		g.Describe("with a valid interchange message", func() {
			var message interchange.DeviceMessage

			device := struct {
				id string
			}{"d1d1d1d1d1d1d1d1d1d1"}

			g.BeforeEach(func() {
				message = interchange.DeviceMessage{
					Authentication: &interchange.DeviceMessageAuthentication{
						DeviceID: device.id,
					},
				}
			})

			g.It("fails when an error is returned during signing", func() {
				scaffold.signer.errors = append(scaffold.signer.errors, fmt.Errorf("bad-sign"))
				e := scaffold.connection.Send(message)
				g.Assert(e.Error()).Equal("bad-sign")
			})

			g.It("fails when an error is returned from the streamer's NextWriter", func() {
				scaffold.streamer.responses = append(scaffold.streamer.responses, testStreamerResponse{
					e: fmt.Errorf("bad-writer"),
				})
				e := scaffold.connection.Send(message)
				g.Assert(e.Error()).Equal("bad-writer")
			})

			g.It("fails when an error is returned from the streamer's NextWriter writer", func() {
				scaffold.streamer.responses = append(scaffold.streamer.responses, testStreamerResponse{
					w: &testWriteCloser{errors: []error{fmt.Errorf("bad-writer")}},
				})
				e := scaffold.connection.Send(message)
				g.Assert(e.Error()).Equal("bad-writer")
			})
		})
	})

	g.Describe("GetID", func() {
		id := uuid.NewV4()
		conn := StreamerConnection{
			LeveledLogger: newStreamerLogger(),
			id:            id,
		}

		g.Assert(conn.GetID()).Equal(id.String())
	})
}
