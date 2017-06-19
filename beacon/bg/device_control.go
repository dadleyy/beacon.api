package bg

import "io"
import "log"
import "sync"
import "time"
import "encoding/json"
import "github.com/gorilla/websocket"

import "github.com/dadleyy/beacon.api/beacon/device"

type DeviceControlProcessor struct {
	*log.Logger

	LogStream     chan<- io.Reader
	ControlStream <-chan io.Reader
	Registrations <-chan *device.Connection

	pool []*device.Connection
}

func (processor *DeviceControlProcessor) handle(message io.Reader, wg *sync.WaitGroup) {
	decoder, command := json.NewDecoder(message), device.ControlMessage{}
	defer wg.Done()

	if e := decoder.Decode(&command); e != nil {
		processor.Printf("received strange control message: %s", e.Error())
		return
	}

	var device *device.Connection

	for _, d := range processor.pool {
		if deviceId := d.UUID.String(); deviceId != command.DeviceId {
			continue
		}

		device = d
		break
	}

	if device == nil {
		processor.Printf("unable to locate device for command, command device id: %s", command.DeviceId)
		return
	}

	writer, e := device.NextWriter(websocket.TextMessage)

	if e != nil {
		processor.Printf("unable to open writer to device (closing device): %s", e.Error())
		processor.unsubscribe(device)
		return
	}

	encoder := json.NewEncoder(writer)

	if e := encoder.Encode(command); e != nil {
		processor.Printf("unable to write command to device (closing device): %s", e.Error())
		processor.unsubscribe(device)
		return
	}

	processor.Printf("relayed command to device[%s] - %v", device.GetID(), command)
}

func (processor *DeviceControlProcessor) unsubscribe(connection *device.Connection) {
	defer connection.Close()
	pool, targetId := make([]*device.Connection, 0, len(processor.pool)-1), connection.UUID.String()

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
	writer, e := connection.NextWriter(websocket.TextMessage)

	if e != nil {
		processor.Printf("unable to get welcome writer for device[%s]: %s", connection.GetID(), e.Error())
		return
	}

	encoder := json.NewEncoder(writer)

	if e := encoder.Encode(device.WelcomeMessage{connection.GetID()}); e != nil {
		processor.Printf("unable to welcome device[%s]: %s", connection.GetID(), e.Error())
		return
	}

	if e := writer.Close(); e != nil {
		processor.Printf("unable to close welcome writer device[%s]: %s", connection.GetID(), e.Error())
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

		decoder, message := json.NewDecoder(reader), struct {
			Data string `json:"data"`
		}{}

		if e := decoder.Decode(&message); e != nil {
			processor.Printf("unable to decode device message: %s", e.Error())
			break
		}

		processor.Printf("recieved message: \"%s\"", message.Data)
	}

	processor.Printf("closing device[%s]", connection.UUID.String())
}

func (processor *DeviceControlProcessor) Start(wg *sync.WaitGroup) {
	defer wg.Done()

	wait, timer := sync.WaitGroup{}, time.NewTicker(time.Minute)
	defer timer.Stop()

	for {
		select {
		case message := <-processor.ControlStream:
			wait.Add(1)
			go processor.handle(message, &wait)
		case connection := <-processor.Registrations:
			wait.Add(2)
			go processor.welcome(connection, &wait)
			go processor.subscribe(connection, &wait)
		case <-timer.C:
			processor.Printf("pool len[%d] cap[%d]", len(processor.pool), cap(processor.pool))
		}
	}

	wait.Wait()
}
