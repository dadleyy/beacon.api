package routes

import "log"
import "fmt"
import "bytes"
import "testing"
import "net/url"
import "net/http/httptest"
import "github.com/franela/goblin"

import "github.com/dadleyy/beacon.api/beacon/net"
import "github.com/dadleyy/beacon.api/beacon/defs"
import "github.com/dadleyy/beacon.api/beacon/logging"
import "github.com/dadleyy/beacon.api/beacon/device"

func newDevicesAPILogger() *logging.Logger {
	out := bytes.NewBuffer([]byte{})
	logger := log.New(out, "", 0)
	logger.SetFlags(0)
	return &logging.Logger{Logger: logger}
}

type testDevicesAPIScaffolding struct {
	api        *Devices
	registry   *testDeviceRegistry
	tokenStore *testDeviceTokenStore
	runtime    *net.RequestRuntime
	body       *bytes.Buffer
	pathValues url.Values
}

func prepareDeviceAPIScaffold() testDevicesAPIScaffolding {
	registry := testDeviceRegistry{}
	tokenStore := testDeviceTokenStore{}
	api := Devices{
		LeveledLogger: newDevicesAPILogger(),
		Registry:      &registry,
		TokenStore:    &tokenStore,
	}

	body := bytes.NewBuffer([]byte{})

	request := httptest.NewRequest("GET", "/device-messages", body)

	pathValues := make(url.Values)

	return testDevicesAPIScaffolding{
		api:        &api,
		registry:   &registry,
		tokenStore: &tokenStore,
		body:       body,
		pathValues: pathValues,
		runtime: &net.RequestRuntime{
			Request: request,
			Values:  pathValues,
		},
	}
}

func Test_DevicesAPI(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("ListDevices", func() {
		var scaffold testDevicesAPIScaffolding

		g.BeforeEach(func() {
			scaffold = prepareDeviceAPIScaffold()
		})

		g.It("errors if unable to get a list of registrations from the registry", func() {
			registry := scaffold.registry
			registry.listRegistrationErrors = append(registry.listRegistrationErrors, fmt.Errorf("bad-list"))
			r := scaffold.api.ListDevices(scaffold.runtime)
			g.Assert(r.Errors[0].Error()).Equal(defs.ErrServerError)
		})

		g.It("returns the list of registered devices if present", func() {
			registry := scaffold.registry
			registry.activeRegistrations = append(registry.activeRegistrations, device.RegistrationDetails{})
			r := scaffold.api.ListDevices(scaffold.runtime)
			g.Assert(len(r.Errors)).Equal(0)
			l, e := r.Results.([]device.RegistrationDetails)
			g.Assert(e).Equal(true)
			g.Assert(len(l)).Equal(1)
		})
	})

	g.Describe("UpdateShorthand", func() {
		var scaffold testDevicesAPIScaffolding

		g.BeforeEach(func() {
			scaffold = prepareDeviceAPIScaffold()
		})

		g.It("returns a not-found error if unable to find the device in the store", func() {
			r := scaffold.api.UpdateShorthand(scaffold.runtime)
			g.Assert(r.Errors[0].Error()).Equal(defs.ErrNotFound)
		})

		g.Describe("having found a device", func() {
			g.BeforeEach(func() {
				testDevice := device.RegistrationDetails{}
				scaffold.registry.activeRegistrations = append(scaffold.registry.activeRegistrations, testDevice)
			})

			g.It("fails without a valid token header", func() {
				r := scaffold.api.UpdateShorthand(scaffold.runtime)
				g.Assert(r.Errors[0].Error()).Equal(defs.ErrNotFound)
			})

			g.It("with a valid token but not authorized", func() {
				scaffold.runtime.Header.Set(defs.APIUserTokenHeader, "some-token")
				scaffold.tokenStore.authorized = false
				r := scaffold.api.UpdateShorthand(scaffold.runtime)
				g.Assert(r.Errors[0].Error()).Equal(defs.ErrNotFound)
			})

			g.Describe("having authorized successfully", func() {

				g.BeforeEach(func() {
					scaffold.runtime.Header.Set(defs.APIUserTokenHeader, "some-token")
					scaffold.tokenStore.authorized = true
				})

				g.It("errors when the color short hand is not present", func() {
					r := scaffold.api.UpdateShorthand(scaffold.runtime)
					g.Assert(r.Errors[0].Error()).Equal(defs.ErrInvalidColorShorthand)
				})

				g.It("errors when the color short hand is not valid", func() {
					scaffold.pathValues.Set("color", "bad-value")
					r := scaffold.api.UpdateShorthand(scaffold.runtime)
					g.Assert(r.Errors[0].Error()).Equal(defs.ErrInvalidColorShorthand)
				})

				g.Describe("with a valid value", func() {
					g.AfterEach(func() {
						r := scaffold.api.UpdateShorthand(scaffold.runtime)
						g.Assert(len(r.Errors)).Equal(0)
					})

					g.It("succeeds when given \"rand\"", func() {
						scaffold.pathValues.Set("color", "rand")
					})

					g.It("succeeds when given \"green\"", func() {
						scaffold.pathValues.Set("color", "green")
					})

					g.It("succeeds when given \"red\"", func() {
						scaffold.pathValues.Set("color", "red")
					})

					g.It("succeeds when given \"off\"", func() {
						scaffold.pathValues.Set("color", "off")
					})

					g.It("succeeds when given \"blue\"", func() {
						scaffold.pathValues.Set("color", "blue")
					})

					g.It("succeeds when given a valid 6 character hex code", func() {
						scaffold.pathValues.Set("color", "ffffff")
					})
				})
			})
		})
	})
}
