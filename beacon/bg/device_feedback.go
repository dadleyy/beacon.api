package bg

import "io"
import "log"
import "sync"

// DeviceFeedbackProcessor is responsible for receiving from the device feedback stream
type DeviceFeedbackProcessor struct {
	*log.Logger
	LogStream <-chan io.Reader
}

// Start is the Processor#Start implementation
func (processor *DeviceFeedbackProcessor) Start(wg *sync.WaitGroup, stop KillSwitch) {
	defer wg.Done()
	running := true

	processor.Printf("device feedback processor starting")

	for running {
		select {
		case <-processor.LogStream:
			processor.Printf("receieved message from device")
		case <-stop:
			processor.Printf("received kill signal, breaking")
			running = false
			break
		}
	}
}
