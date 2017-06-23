package bg

import "os"
import "io"
import "log"
import "sync"
import "time"
import "io/ioutil"

import "github.com/gorilla/websocket"
import "github.com/golang/protobuf/proto"

import "github.com/dadleyy/beacon.api/beacon/device"
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
func NewDeviceControlProcessor(channels *DeviceChannels, store device.Index) *DeviceControlProcessor {
	logger := log.New(os.Stdout, "device control ", log.Ldate|log.Ltime|log.Lshortfile)
	var pool []*device.Connection
	return &DeviceControlProcessor{logger, channels, store, pool}
}

// The DeviceControlProcessor is used by the server to maintain the pool of websocket connections, register new device
// connections w/ the index and relay any messages along to the device.
type DeviceControlProcessor struct {
	*log.Logger
	channels *DeviceChannels
	index    device.Index
	pool     []*device.Connection
}

func (processor *DeviceControlProcessor) handle(message io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()

	messageData, e := ioutil.ReadAll(message)

	if e != nil {
		processor.Printf("unable to read message: %s", e.Error())
		return
	}

	controlMessage := interchange.DeviceMessage{}

	if e := proto.Unmarshal(messageData, &controlMessage); e != nil {
		processor.Printf("unable to unmarshal message: %s", e.Error())
		return
	}

	var device *device.Connection
	targetID := controlMessage.Authentication.DeviceID

	for _, d := range processor.pool {
		if deviceID := d.UUID.String(); deviceID != targetID {
			continue
		}

		device = d
		break
	}

	if device == nil {
		processor.Printf("unable to locate device for command, command device id: %s", targetID)
		return
	}

	writer, e := device.NextWriter(websocket.TextMessage)

	if e != nil {
		processor.Printf("unable to open writer to device (closing device): %s", e.Error())
		processor.unsubscribe(device)
		return
	}

	defer writer.Close()

	data, e := proto.Marshal(&controlMessage)

	if e != nil {
		processor.Printf("unable to write command to device (closing device): %s", e.Error())
		return
	}

	if _, e := writer.Write(data); e != nil {
		processor.Printf("unable to write command to device (closing device): %s", e.Error())
		processor.unsubscribe(device)
		return
	}

	processor.Printf("relayed command to device[%s]", device.GetID())
}

func (processor *DeviceControlProcessor) unsubscribe(connection *device.Connection) {
	defer connection.Close()
	pool, targetID := make([]*device.Connection, 0, len(processor.pool)-1), connection.UUID.String()

	if e := processor.index.Remove(targetID); e != nil {
		processor.Printf("[warn] unable to get current list of devices - %s", e.Error())
	}

	for _, device := range processor.pool {
		if deviceID := device.UUID.String(); deviceID == targetID {
			continue
		}

		pool = append(pool, device)
	}

	processor.pool = pool
}

func (processor *DeviceControlProcessor) welcome(connection *device.Connection, wg *sync.WaitGroup) {
	defer wg.Done()
	writer, e := connection.NextWriter(websocket.TextMessage)

	if e != nil {
		processor.Printf("unable to get welcome writer for device[%s]: %s", connection.GetID(), e.Error())
		return
	}

	defer writer.Close()

	welcomeData, e := proto.Marshal(&interchange.WelcomeMessage{
		DeviceID: connection.GetID(),
		Body:     "Hello world, I am the body of the welcome message!",
	})

	if e != nil {
		processor.Printf("unable to welcome device[%s]: %s", connection.GetID(), e.Error())
		return
	}

	welcomeMessage := interchange.DeviceMessage{
		Type: interchange.DeviceMessageType_WELCOME,
		Authentication: &interchange.DeviceMessageAuthentication{
			DeviceID: connection.GetID(),
		},
		Payload: welcomeData,
	}

	messageData, e := proto.Marshal(&welcomeMessage)

	if e != nil {
		processor.Printf("unable to welcome device[%s]: %s", connection.GetID(), e.Error())
		return
	}

	if _, e := writer.Write(messageData); e != nil {
		processor.Printf("unable to push device id into store: %s", e.Error())
		return
	}

	processor.Printf("welcomed device[%s]", connection.GetID())
}

func (processor *DeviceControlProcessor) subscribe(connection *device.Connection, wg *sync.WaitGroup) {
	defer wg.Done()
	defer processor.unsubscribe(connection)
	processor.pool = append(processor.pool, connection)

	processor.Printf("subscribing to device[%s]", connection.UUID.String())

	for {
		messageType, reader, e := connection.NextReader()

		if e != nil {
			processor.Printf("unable to read from device: %s", e.Error())
			break
		}

		if messageType != websocket.TextMessage {
			processor.Printf("received strange message from device, closing connection")
			break
		}

		processor.channels.Feedback <- reader
	}

	processor.Printf("closing device[%s]", connection.UUID.String())
}

// Start will continously loop over registration & command channels delegating to private methods as necessary.
func (processor *DeviceControlProcessor) Start(wg *sync.WaitGroup, stop KillSwitch) {
	defer wg.Done()

	processor.Printf("device control processor starting")

	wait, timer, running := sync.WaitGroup{}, time.NewTicker(time.Minute), true
	defer timer.Stop()

	for running {
		select {
		case message := <-processor.channels.Commands:
			wait.Add(1)
			processor.Printf("received message on read channel")
			go processor.handle(message, &wait)
		case connection := <-processor.channels.Registrations:
			wait.Add(2)
			go processor.welcome(connection, &wait)
			go processor.subscribe(connection, &wait)
		case <-timer.C:
			processor.Printf("pool len[%d] cap[%d]", len(processor.pool), cap(processor.pool))
		case <-stop:
			processor.Printf("received kill signal, breaking")
			running = false
			break
		}
	}

	for _, c := range processor.pool {
		processor.Printf("closing connection: %s", c.GetID())
		c.Close()
	}

	wait.Wait()
}
