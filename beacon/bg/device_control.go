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

type WriteStream chan<- io.Reader
type ReadStream <-chan io.Reader
type RegistrationStream <-chan *device.Connection

type DeviceChannels struct {
	Commands      ReadStream
	Feedback      WriteStream
	Registrations RegistrationStream
}

func NewDeviceControlProcessor(channels *DeviceChannels, store device.Registry) *DeviceControlProcessor {
	logger := log.New(os.Stdout, "device control ", log.Ldate|log.Ltime|log.Lshortfile)
	pool := make([]*device.Connection, 0)
	return &DeviceControlProcessor{logger, channels, store, pool}
}

type DeviceControlProcessor struct {
	*log.Logger
	channels *DeviceChannels
	registry device.Registry
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
	targetId := controlMessage.Authentication.DeviceId

	for _, d := range processor.pool {
		if deviceId := d.UUID.String(); deviceId != targetId {
			continue
		}

		device = d
		break
	}

	if device == nil {
		processor.Printf("unable to locate device for command, command device id: %s", targetId)
		return
	}

	writer, e := device.NextWriter(websocket.BinaryMessage)

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
	pool, targetId := make([]*device.Connection, 0, len(processor.pool)-1), connection.UUID.String()

	if e := processor.registry.Remove(targetId); e != nil {
		processor.Printf("[warn] unable to get current list of devices - %s", e.Error())
	}

	for _, device := range processor.pool {
		if deviceId := device.UUID.String(); deviceId == targetId {
			continue
		}

		pool = append(pool, device)
	}

	processor.pool = pool
}

func (processor *DeviceControlProcessor) welcome(connection *device.Connection, wg *sync.WaitGroup) {
	defer wg.Done()
	writer, e := connection.NextWriter(websocket.BinaryMessage)

	if e != nil {
		processor.Printf("unable to get welcome writer for device[%s]: %s", connection.GetID(), e.Error())
		return
	}

	defer writer.Close()

	welcomeMessage := interchange.DeviceMessage{
		Authentication: &interchange.DeviceMessageAuthentication{
			DeviceId: connection.GetID(),
		},
		RequestBody: &interchange.DeviceMessage_Welcome{&interchange.WelcomeMessage{}},
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

	if e := processor.registry.Insert(connection.GetID()); e != nil {
		processor.Printf("unable to push device id into store: %s", e.Error())
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
