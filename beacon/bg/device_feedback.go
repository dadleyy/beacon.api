package bg

import "io"
import "sync"

import "github.com/dadleyy/beacon.api/beacon/defs"
import "github.com/dadleyy/beacon.api/beacon/logging"

// NewDeviceFeedbackProcessor is responsible for receiving from the device feedback stream
func NewDeviceFeedbackProcessor(feedback ReadStream) *DeviceFeedbackProcessor {
	logger := logging.New(defs.DeviceFeedbackLogPrefix, logging.Cyan)
	return &DeviceFeedbackProcessor{logger, feedback}
}

// DeviceFeedbackProcessor is responsible for receiving from the device feedback stream
type DeviceFeedbackProcessor struct {
	*logging.Logger
	feedback <-chan io.Reader
}

// Start is the Processor#Start implementation
func (processor *DeviceFeedbackProcessor) Start(wg *sync.WaitGroup, stop KillSwitch) {
	defer wg.Done()
	running := true

	processor.Infof("device feedback processor starting")

	for running {
		select {
		case _, ok := <-processor.feedback:
			if ok != true {
				return
			}

			processor.Debugf("receieved message from device")
		case <-stop:
			processor.Warnf("received kill signal, breaking")
			running = false
			break
		}
	}
}
