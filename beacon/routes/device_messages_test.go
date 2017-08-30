package routes

import "log"
import "fmt"
import "bytes"
import "testing"
import "net/http/httptest"

import "github.com/franela/goblin"
import "github.com/dadleyy/beacon.api/beacon/net"
import "github.com/dadleyy/beacon.api/beacon/defs"
import "github.com/dadleyy/beacon.api/beacon/device"
import "github.com/dadleyy/beacon.api/beacon/logging"

func newDeviceMessagesAPILogger() *logging.Logger {
	out := bytes.NewBuffer([]byte{})
	logger := log.New(out, "", 0)
	logger.SetFlags(0)
	return &logging.Logger{Logger: logger}
}

type testDeviceMessagesAPIScaffolding struct {
	api       *DeviceMessages
	internals *testDeviceMessagesAPIInternals
	runtime   *net.RequestRuntime
	body      *bytes.Buffer
}

type testDeviceMessagesAPIInternals struct {
	authorized    bool
	createdTokens []device.TokenDetails
	foundTokens   []device.TokenDetails
	foundDevices  []device.RegistrationDetails
	removalErrors []error
}

func (t *testDeviceMessagesAPIInternals) RemoveDevice(string) error {
	if len(t.removalErrors) >= 1 {
		return t.removalErrors[0]
	}

	return nil
}

func (t *testDeviceMessagesAPIInternals) FindDevice(string) (device.RegistrationDetails, error) {
	if len(t.foundDevices) >= 1 {
		return t.foundDevices[0], nil
	}

	return device.RegistrationDetails{}, fmt.Errorf("not-found")
}

func (t *testDeviceMessagesAPIInternals) CreateToken(string, string, uint) (device.TokenDetails, error) {
	if len(t.createdTokens) >= 1 {
		return t.createdTokens[0], nil
	}

	return device.TokenDetails{}, fmt.Errorf("not-found")
}

func (t *testDeviceMessagesAPIInternals) ListTokens(string) ([]device.TokenDetails, error) {
	if len(t.foundTokens) >= 1 {
		return t.foundTokens, nil
	}

	return nil, fmt.Errorf("not-found")
}

func (t *testDeviceMessagesAPIInternals) AuthorizeToken(string, string, uint) bool {
	return t.authorized
}

func Test_DeviceMessagesAPI(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("CreateMessage", func() {
		var scaffold testDeviceMessagesAPIScaffolding

		g.BeforeEach(func() {
			internals := &testDeviceMessagesAPIInternals{
				createdTokens: make([]device.TokenDetails, 0),
				foundTokens:   make([]device.TokenDetails, 0),
				foundDevices:  make([]device.RegistrationDetails, 0),
				removalErrors: make([]error, 0),
			}

			api := &DeviceMessages{
				LeveledLogger: newDeviceMessagesAPILogger(),
				TokenStore:    internals,
				Index:         internals,
			}

			body := bytes.NewBuffer([]byte{})

			request := httptest.NewRequest("GET", "/device-messages", body)

			scaffold = testDeviceMessagesAPIScaffolding{
				api:       api,
				internals: internals,
				body:      body,
				runtime: &net.RequestRuntime{
					Request: request,
				},
			}
		})

		g.It("fails if it is unable to read the body of the request reasonably", func() {
			r := scaffold.api.CreateMessage(scaffold.runtime)
			g.Assert(r.Errors[0].Error()).Equal(defs.ErrBadRequestFormat)
		})

		g.Describe("with a valid json body", func() {
			g.BeforeEach(func() {
				scaffold.body.Write([]byte("{\"device_id\": \"123\"}"))
			})

			g.It("fails when unable to find the device", func() {
				r := scaffold.api.CreateMessage(scaffold.runtime)
				g.Assert(r.Errors[0].Error()).Equal(defs.ErrNotFound)
			})

			g.Describe("when a device was found successfully", func() {
				device := device.RegistrationDetails{}

				g.BeforeEach(func() {
					scaffold.internals.foundDevices = append(scaffold.internals.foundDevices, device)
				})

				g.It("should fail when no authorization header was present", func() {
					r := scaffold.api.CreateMessage(scaffold.runtime)
					g.Assert(r.Errors[0].Error()).Equal(defs.ErrNotFound)
				})

				g.It("fails even with header but unable to auth header", func() {
					scaffold.runtime.Header.Set(defs.APIUserTokenHeader, "some-token")
					r := scaffold.api.CreateMessage(scaffold.runtime)
					g.Assert(r.Errors[0].Error()).Equal(defs.ErrNotFound)
				})

				g.It("succeeds if authorized w/ valid body", func() {
					scaffold.internals.authorized = true
					scaffold.runtime.Header.Set(defs.APIUserTokenHeader, "some-token")
					r := scaffold.api.CreateMessage(scaffold.runtime)
					g.Assert(len(r.Errors)).Equal(0)
				})
			})
		})

	})
}
