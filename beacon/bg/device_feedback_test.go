package bg

import "io"
import "sync"
import "bytes"
import "strings"
import "testing"
import "github.com/franela/goblin"

type deviceFeedbackScaffold struct {
	receiver  chan io.Reader
	wg        *sync.WaitGroup
	kill      KillSwitch
	processor *DeviceFeedbackProcessor
	log       *bytes.Buffer
}

func (s *deviceFeedbackScaffold) Reset() {
	s.receiver = make(chan io.Reader)
	s.kill = make(KillSwitch, 1)
	s.wg = &sync.WaitGroup{}
	s.log = bytes.NewBuffer([]byte{})
	s.processor = &DeviceFeedbackProcessor{
		Logger:   newTestLogger(s.log),
		feedback: s.receiver,
	}
}

func Test_DeviceFeedback(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("DeviceFeedback", func() {
		s := &deviceFeedbackScaffold{}

		g.BeforeEach(s.Reset)

		g.It("successfully terminates after having received all the feedback items", func() {
			s.wg.Add(1)
			go s.processor.Start(s.wg, s.kill)
			s.receiver <- bytes.NewBuffer([]byte{})
			close(s.receiver)
			s.wg.Wait()
		})

		g.It("successfully terminates when kill signal is given", func() {
			s.wg.Add(1)
			go s.processor.Start(s.wg, s.kill)
			g.Assert(strings.Contains(s.log.String(), "kill signal")).Equal(false)
			s.kill <- struct{}{}
			s.wg.Wait()
			g.Assert(strings.Contains(s.log.String(), "kill signal")).Equal(true)
		})

	})
}
