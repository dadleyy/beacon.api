package bg

import "io"
import "fmt"
import "log"
import "sync"
import "bytes"
import "strings"
import "testing"
import "crypto/rsa"
import "crypto/rand"
import "github.com/franela/goblin"
import "github.com/golang/protobuf/proto"
import "github.com/dadleyy/beacon.api/beacon/device"
import "github.com/dadleyy/beacon.api/beacon/logging"
import "github.com/dadleyy/beacon.api/beacon/security"
import "github.com/dadleyy/beacon.api/beacon/interchange"

func newTestLogger(output io.Writer) *logging.Logger {
	logger := log.New(output, "", 0)
	logger.SetFlags(0)
	return &logging.Logger{Logger: logger}
}

type deviceControlScaffold struct {
	key           *security.ServerKey
	log           *bytes.Buffer
	connections   []device.Connection
	index         *testDeviceIndex
	channels      []chan io.Reader
	registrations device.RegistrationStream
	processor     *DeviceControlProcessor
	wg            *sync.WaitGroup
	kill          KillSwitch
}

func (s *deviceControlScaffold) Reset() {
	s.connections = make([]device.Connection, 0)

	s.log = bytes.NewBuffer([]byte{})

	s.index = &testDeviceIndex{}

	s.channels = []chan io.Reader{
		make(chan io.Reader, 1),
		make(chan io.Reader, 1),
	}

	k, e := rsa.GenerateKey(rand.Reader, 1024)

	if e != nil {
		panic(e)
	}

	s.key = &security.ServerKey{
		PrivateKey: k,
	}

	s.registrations = make(device.RegistrationStream, 1)

	s.processor = &DeviceControlProcessor{
		Logger: newTestLogger(s.log),
		key:    s.key,
		channels: &DeviceChannels{
			Commands:      s.channels[0],
			Feedback:      s.channels[1],
			Registrations: s.registrations,
		},
		index: s.index,
		pool:  s.connections,
	}

	s.wg = &sync.WaitGroup{}

	s.kill = make(KillSwitch)
}

func (s *deviceControlScaffold) sendKillSignal() {
	s.kill <- struct{}{}
}

type lastErrorLister struct {
}

func (l lastErrorLister) lastError(list []error) error {
	if len(list) >= 1 {
		return list[0]
	}
	return nil
}

func (l lastErrorLister) lastErrorOrNotFound(list []error) error {
	e := l.lastError(list)

	if e != nil {
		return e
	}

	return fmt.Errorf("not-found")
}

type testDeviceIndex struct {
	lastErrorLister
	errors  []error
	devices []device.RegistrationDetails
}

func (i testDeviceIndex) RemoveDevice(string) error {
	return i.lastError(i.errors)
}

func (i testDeviceIndex) FindDevice(string) (device.RegistrationDetails, error) {
	if len(i.devices) >= 1 {
		return i.devices[0], nil
	}

	return device.RegistrationDetails{}, i.lastErrorOrNotFound(i.errors)
}

type testConnection struct {
	lastErrorLister
	closed       bool
	id           string
	sentMessages []interchange.DeviceMessage
	readers      []io.Reader
	errors       []error
}

func (c *testConnection) GetID() string {
	return c.id
}

func (c *testConnection) Send(m interchange.DeviceMessage) error {
	if c.sentMessages == nil {
		c.sentMessages = make([]interchange.DeviceMessage, 0)
	}

	c.sentMessages = append(c.sentMessages, m)

	return c.lastError(c.errors)
}

func (c *testConnection) Receive() (io.Reader, error) {
	if len(c.readers) >= 1 {
		r := c.readers[0]
		c.readers = c.readers[1:]
		return r, nil
	}

	return nil, c.lastErrorOrNotFound(c.errors)
}

func (c *testConnection) Close() error {
	c.closed = true
	return nil
}

type testReader struct {
	lastErrorLister
	errors []error
}

func (r *testReader) Read([]byte) (int, error) {
	return 0, r.lastError(r.errors)
}

func Test_DeviceControl(t *testing.T) {
	g := goblin.Goblin(t)

	scaffold := &deviceControlScaffold{}

	g.Describe("DeviceControl", func() {

		g.BeforeEach(scaffold.Reset)

		g.Describe("#subscribe", func() {
			var connection *testConnection
			var wg *sync.WaitGroup

			g.BeforeEach(func() {
				connection = &testConnection{}
				wg = &sync.WaitGroup{}
			})

			g.It("returns the error sent from the connection during an attempt to receive", func() {
				wg.Add(1)
				connection.errors = append(connection.errors, fmt.Errorf("bad-receive"))
				e := scaffold.processor.subscribe(connection, wg)
				wg.Wait()
				g.Assert(e.Error()).Equal("bad-receive")
			})

			g.It("sends the feedback message on a successful receive to the feedback channel", func() {
				wg.Add(1)
				connection.readers = append(connection.readers, bytes.NewBuffer([]byte("hello world")))
				scaffold.processor.subscribe(connection, wg)
				wg.Wait()
				feedback := <-scaffold.channels[1]
				reader, ok := feedback.(*bytes.Buffer)
				g.Assert(ok).Equal(true)
				g.Assert(reader.String()).Equal("hello world")
			})
		})

		g.Describe("#unsubscribe", func() {
			var connection *testConnection

			g.BeforeEach(func() {
				connection = &testConnection{id: "patriots"}

				scaffold.processor.pool = []device.Connection{
					&testConnection{id: "buffalo"},
					&testConnection{id: "bills"},
				}

				scaffold.processor.pool = append(scaffold.processor.pool, connection)
			})

			g.It("returns the error returned from the index if unable to remove", func() {
				scaffold.index.errors = []error{
					fmt.Errorf("bad-remove"),
				}
				e := scaffold.processor.unsubscribe(connection)
				g.Assert(e.Error()).Equal("bad-remove")
			})

			g.It("removes the device if no problems removing from the index", func() {
				g.Assert(len(scaffold.processor.pool)).Equal(3)
				e := scaffold.processor.unsubscribe(connection)
				g.Assert(e).Equal(nil)
				g.Assert(len(scaffold.processor.pool)).Equal(2)
			})

		})

		g.Describe("#Start", func() {

			g.BeforeEach(func() {
				scaffold.wg.Add(1)
			})

			g.AfterEach(func() {
				scaffold.wg.Wait()
			})

			g.Describe("receiving registrations", func() {

				g.It("sends a welcome message to the device", func() {
					connection := &testConnection{
						id: "some-device",
					}
					scaffold.registrations <- connection
					g.Assert(len(connection.sentMessages)).Equal(0)
					go scaffold.processor.Start(scaffold.wg, scaffold.kill)
					close(scaffold.registrations)
					scaffold.wg.Wait()
					g.Assert(len(connection.sentMessages)).Equal(1)
					payload, welcome := connection.sentMessages[0].GetPayload(), interchange.WelcomeMessage{}
					err := proto.Unmarshal(payload, &welcome)
					g.Assert(err).Equal(nil)
					g.Assert(welcome.GetDeviceID()).Equal("some-device")
				})

				g.It("logs any errors that come out of the connection's message delivery", func() {
					connection := &testConnection{
						errors: []error{fmt.Errorf("bad-welcome-send")},
					}
					scaffold.registrations <- connection
					g.Assert(strings.Contains(scaffold.log.String(), "bad-welcome-send")).Equal(false)
					go scaffold.processor.Start(scaffold.wg, scaffold.kill)
					close(scaffold.registrations)
					scaffold.wg.Wait()
					g.Assert(strings.Contains(scaffold.log.String(), "bad-welcome-send")).Equal(true)
				})

			})

			g.Describe("receieving commands", func() {

				g.It("logs any error during read from the reader sent into the channel", func() {
					errorString := "bad-read"
					scaffold.channels[0] <- &testReader{
						errors: []error{fmt.Errorf(errorString)},
					}

					found := strings.Contains(scaffold.log.String(), errorString)
					g.Assert(found).Equal(false)

					go scaffold.processor.Start(scaffold.wg, scaffold.kill)
					close(scaffold.channels[0])
					scaffold.wg.Wait()

					found = strings.Contains(scaffold.log.String(), errorString)
					g.Assert(found).Equal(true)
				})

				g.It("logs any error during unmarshalling of received message", func() {
					scaffold.channels[0] <- bytes.NewBuffer([]byte("dasdasd{}{}{}"))
					g.Assert(strings.Contains(scaffold.log.String(), "unmarshal")).Equal(false)
					go scaffold.processor.Start(scaffold.wg, scaffold.kill)
					close(scaffold.channels[0])
					scaffold.wg.Wait()
					g.Assert(strings.Contains(scaffold.log.String(), "unmarshal")).Equal(true)
				})

				g.Describe("having been given a valid control message", func() {
					g.BeforeEach(func() {
						b, _ := proto.Marshal(&interchange.DeviceMessage{
							Authentication: &interchange.DeviceMessageAuthentication{
								DeviceID: "some-device",
							},
						})
						scaffold.channels[0] <- bytes.NewBuffer(b)
					})

					g.It("logs it's inability to find a device if none are found in the index", func() {
						g.Assert(strings.Contains(scaffold.log.String(), "unable to locate")).Equal(false)
						go scaffold.processor.Start(scaffold.wg, scaffold.kill)
						close(scaffold.channels[0])
						scaffold.wg.Wait()
						g.Assert(strings.Contains(scaffold.log.String(), "unable to locate")).Equal(true)
					})

					g.It("sends the command to the device if it is able to find it", func() {
						connection := &testConnection{
							id: "some-device",
						}
						scaffold.processor.pool = append(scaffold.processor.pool, connection)
						g.Assert(len(connection.sentMessages)).Equal(0)
						go scaffold.processor.Start(scaffold.wg, scaffold.kill)
						close(scaffold.channels[0])
						scaffold.wg.Wait()
						g.Assert(len(connection.sentMessages)).Equal(1)
					})

					g.It("logs the error returned from the device, if exists", func() {
						connection := &testConnection{
							id:     "some-device",
							errors: []error{fmt.Errorf("some-bad-write")},
						}
						scaffold.processor.pool = append(scaffold.processor.pool, connection)
						g.Assert(len(connection.sentMessages)).Equal(0)
						g.Assert(strings.Contains(scaffold.log.String(), "some-bad-write")).Equal(false)
						go scaffold.processor.Start(scaffold.wg, scaffold.kill)
						close(scaffold.channels[0])
						scaffold.wg.Wait()
						g.Assert(len(connection.sentMessages)).Equal(1)
						g.Assert(strings.Contains(scaffold.log.String(), "some-bad-write")).Equal(true)
					})

				})

				g.It("immediately stops when the command stream channel is closed", func() {
					connection := &testConnection{}
					scaffold.processor.pool = append(scaffold.processor.pool, connection)
					close(scaffold.channels[0])
					g.Assert(connection.closed).Equal(false)
					scaffold.processor.Start(scaffold.wg, scaffold.kill)
					scaffold.wg.Wait()
					g.Assert(connection.closed).Equal(true)
				})

				g.It("immediately stops when the registration stream channel is closed", func() {
					connection := &testConnection{}
					scaffold.processor.pool = append(scaffold.processor.pool, connection)
					close(scaffold.registrations)
					g.Assert(connection.closed).Equal(false)
					scaffold.processor.Start(scaffold.wg, scaffold.kill)
					scaffold.wg.Wait()
					g.Assert(connection.closed).Equal(true)
				})

			})

			g.Describe("when kill switch is sent", func() {
				g.BeforeEach(func() {
					go scaffold.sendKillSignal()
				})

				g.It("closes any connections in the pool when kill switch is sent", func() {
					connection := &testConnection{}
					scaffold.processor.pool = append(scaffold.processor.pool, connection)
					g.Assert(connection.closed).Equal(false)
					scaffold.processor.Start(scaffold.wg, scaffold.kill)
					scaffold.wg.Wait()
					g.Assert(connection.closed).Equal(true)
				})
			})

		})

	})
}
