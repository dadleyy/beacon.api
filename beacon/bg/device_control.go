package bg

import "io"
import "sync"
import "time"
import "io/ioutil"

import "github.com/golang/protobuf/proto"

import "github.com/dadleyy/beacon.api/beacon/defs"
import "github.com/dadleyy/beacon.api/beacon/device"
import "github.com/dadleyy/beacon.api/beacon/logging"
import "github.com/dadleyy/beacon.api/beacon/security"
import "github.com/dadleyy/beacon.api/beacon/interchange"

// WriteStream defines a send-only channel for io.Reader types
type WriteStream chan<- io.Reader

// ReadStream defines a receive-only channel for io.Reader types
type ReadStream <-chan io.Reader

// DeviceChannels is a convenience structure containing a ReadStream, WriteStream and RegistrationStream
type DeviceChannels struct {
	Commands      ReadStream
	Feedback      WriteStream
	Registrations device.RegistrationStream
}

// NewDeviceControlProcessor returns a new DeviceControlProcessor
func NewDeviceControlProcessor(channels *DeviceChannels, store device.Index, key *security.ServerKey) *DeviceControlProcessor {
	logger := logging.New(defs.DeviceControlLogPrefix, logging.Yellow)
	var pool []device.Connection
	return &DeviceControlProcessor{logger, key, channels, store, pool}
}

// The DeviceControlProcessor is used by the server to maintain the pool of websocket connections, register new device
// connections w/ the index and relay any messages along to the device.
type DeviceControlProcessor struct {
	*logging.Logger
	key      *security.ServerKey
	channels *DeviceChannels
	index    device.Index
	pool     []device.Connection
}

// handle receives a reader interface that contains a serialized device message and attempts
func (processor *DeviceControlProcessor) handle(message io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()

	messageData, e := ioutil.ReadAll(message)

	if e != nil {
		processor.Infof("unable to read message: %s", e.Error())
		return
	}

	controlMessage := interchange.DeviceMessage{}

	if e := proto.Unmarshal(messageData, &controlMessage); e != nil {
		processor.Infof("unable to unmarshal message: %s", e.Error())
		return
	}

	var device device.Connection
	targetID := controlMessage.GetAuthentication().GetDeviceID()

	for _, d := range processor.pool {
		processor.Infof("comparing d[%s]", d.GetID())

		if deviceID := d.GetID(); deviceID != targetID {
			continue
		}

		device = d
		break
	}

	if device == nil {
		processor.Warnf("unable to locate device for command, command device id: %s", targetID)
		return
	}

	// At this point we've found a device to send to, write our message into it.
	if e := device.Send(controlMessage); e != nil {
		processor.Warnf("unable to write command to device (closing device): %s", e.Error())
		processor.unsubscribe(device)
		return
	}

	processor.Infof("relayed command to device[%s]", device.GetID())
}

func (processor *DeviceControlProcessor) unsubscribe(connection device.Connection) {
	defer connection.Close()
	pool, targetID := make([]device.Connection, 0, len(processor.pool)-1), connection.GetID()

	if e := processor.index.RemoveDevice(targetID); e != nil {
		processor.Errorf("unable to remove target from device index: %s", e.Error())
		return
	}

	for _, device := range processor.pool {
		if deviceID := device.GetID(); deviceID == targetID {
			continue
		}

		pool = append(pool, device)
	}

	processor.pool = pool
}

func (processor *DeviceControlProcessor) welcome(connection device.Connection, wg *sync.WaitGroup) {
	defer wg.Done()
	secret, e := processor.key.SharedSecret()

	if e != nil {
		processor.Errorf("unable to generate shared secret: %s", e.Error())
		return
	}

	welcomeData, e := proto.Marshal(&interchange.WelcomeMessage{
		DeviceID:     connection.GetID(),
		Body:         defs.WelcomeMessageBody,
		SharedSecret: secret,
	})

	if e != nil {
		processor.Errorf("unable to welcome device[%s]: %s", connection.GetID(), e.Error())
		return
	}

	welcomeMessage := interchange.DeviceMessage{
		Type: interchange.DeviceMessageType_WELCOME,
		Authentication: &interchange.DeviceMessageAuthentication{
			DeviceID: connection.GetID(),
		},
		Payload: welcomeData,
	}

	if e := connection.Send(welcomeMessage); e != nil {
		processor.Warnf("unable to send welcome message: %s", e.Error())
		return
	}

	processor.Infof("welcomed device[%s]", connection.GetID())
}

func (processor *DeviceControlProcessor) subscribe(connection device.Connection, wg *sync.WaitGroup) {
	defer wg.Done()
	defer processor.unsubscribe(connection)
	processor.pool = append(processor.pool, connection)
	processor.Infof("subscribing to device[%s]", connection.GetID())
	connected := true

	for connected {
		reader, e := connection.Receive()

		if e != nil {
			connected = false
			processor.Infof("unable to read from device: %s", e.Error())
			break
		}

		processor.channels.Feedback <- reader
	}

	processor.Infof("closing device[%s]", connection.GetID())
}

// Start will continuously loop over registration & command channels delegating to private methods as necessary.
func (processor *DeviceControlProcessor) Start(wg *sync.WaitGroup, stop KillSwitch) {
	defer wg.Done()

	processor.Infof("device control processor starting")

	wait, timer, running := sync.WaitGroup{}, time.NewTicker(time.Minute), true
	defer timer.Stop()

	for running {
		select {
		case message, ok := <-processor.channels.Commands:
			if !ok {
				running = false
				break
			}

			wait.Add(1)
			processor.Infof("received message on read channel")
			go processor.handle(message, &wait)
		case connection, ok := <-processor.channels.Registrations:
			if ok != true {
				running = false
				break
			}

			wait.Add(2)
			go processor.welcome(connection, &wait)
			go processor.subscribe(connection, &wait)
		case <-timer.C:
			processor.Infof("pool len[%d] cap[%d]", len(processor.pool), cap(processor.pool))
		case <-stop:
			processor.Infof("received kill signal, breaking")
			running = false
			break
		}
	}

	for _, c := range processor.pool {
		processor.Infof("closing connection: %s", c.GetID())
		c.Close()
	}

	wait.Wait()
}
