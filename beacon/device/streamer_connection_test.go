package device

import "log"
import "bytes"
import "testing"
import "github.com/franela/goblin"
import "github.com/satori/go.uuid"
import "github.com/dadleyy/beacon.api/beacon/logging"

func newStreamerLogger() *logging.Logger {
	out := bytes.NewBuffer([]byte{})
	logger := log.New(out, "", 0)
	logger.SetFlags(0)
	return &logging.Logger{Logger: logger}
}

func Test_StreamerConnection(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("Send", func() {
		id := uuid.NewV4()
		conn := StreamerConnection{
			LeveledLogger: newStreamerLogger(),
			id:            id,
		}

		g.Assert(conn.GetID()).Equal(id.String())
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
