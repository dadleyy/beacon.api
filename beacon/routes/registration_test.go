package routes

import "fmt"
import "bytes"
import "testing"
import "encoding/hex"
import "net/http/httptest"

import "github.com/franela/goblin"
import "github.com/dadleyy/beacon.api/beacon/net"
import "github.com/dadleyy/beacon.api/beacon/defs"
import "github.com/dadleyy/beacon.api/beacon/device"

type registrationAPIScaffolding struct {
	api      *RegistrationAPI
	registry *testDeviceRegistry
	runtime  *net.RequestRuntime
	body     *bytes.Buffer
	stream   device.RegistrationStream
}

func prepareRegistrationAPIScaffolding() registrationAPIScaffolding {
	registry := testDeviceRegistry{}
	stream := make(device.RegistrationStream, 0)

	api := RegistrationAPI{
		LeveledLogger: newTestRouteLogger(),
		Registry:      &registry,
		stream:        stream,
	}

	body := bytes.NewBuffer([]byte{})

	runtime := net.RequestRuntime{
		Request: httptest.NewRequest("GET", "/registrations", body),
	}

	return registrationAPIScaffolding{
		api:      &api,
		registry: &registry,
		runtime:  &runtime,
		stream:   stream,
		body:     body,
	}
}

func Test_RegistrationAPI(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("Preregister", func() {
		var scaffold registrationAPIScaffolding

		g.BeforeEach(func() {
			scaffold = prepareRegistrationAPIScaffolding()
		})

		g.It("errors when unable to read request body into a registration req", func() {
			r := scaffold.api.Preregister(scaffold.runtime)
			g.Assert(r.Errors[0].Error()).Equal(defs.ErrBadRequestFormat)
		})

		g.It("errors with an invalid name", func() {
			scaffold.body.Write([]byte(`
			{
				"name": "",
				"shared_secret": "er12er12er12er12er12er12er12er12er12er12"
			}
			`))
			r := scaffold.api.Preregister(scaffold.runtime)
			g.Assert(r.Errors[0].Error()).Equal(defs.ErrBadRequestFormat)
		})

		g.It("errors with an empty object", func() {
			scaffold.body.Write([]byte(`{}`))
			r := scaffold.api.Preregister(scaffold.runtime)
			g.Assert(r.Errors[0].Error()).Equal(defs.ErrBadRequestFormat)
		})

		g.It("errors with an invalid secret", func() {
			scaffold.body.Write([]byte(`
			{
				"name": "some-device",
				"shared_secret": ""
			}
			`))
			r := scaffold.api.Preregister(scaffold.runtime)
			g.Assert(r.Errors[0].Error()).Equal(defs.ErrBadRequestFormat)
		})

		g.Describe("with a valid request body but invalid shared secret", func() {
			g.BeforeEach(func() {
				scaffold.body.Write([]byte(`
				{
					"name": "some-device",
					"shared_secret": "r12r12r12r12r12r12r12r12r12r12"
				}
				`))
			})

			g.It("fails if able to find a device by the same name", func() {
				scaffold.registry.activeRegistrations = append(scaffold.registry.activeRegistrations, device.RegistrationDetails{})
				r := scaffold.api.Preregister(scaffold.runtime)
				g.Assert(r.Errors[0].Error()).Equal(defs.ErrDuplicateRegistrationName)
			})

			g.It("fails if unable to parse the shared secret", func() {
				r := scaffold.api.Preregister(scaffold.runtime)
				g.Assert(r.Errors[0].Error()).Equal(defs.ErrInvalidDeviceSharedSecret)
			})
		})

		g.Describe("with a valid request body but an invalid rsa shared secret", func() {
			g.BeforeEach(func() {
				secretValue := hex.EncodeToString([]byte("a-very-long-shared-secret"))

				body := []byte(fmt.Sprintf(`
				{
					"name": "some-device",
					"shared_secret": "%s"
				}
				`, secretValue))

				scaffold.body.Write(body)
			})

			g.It("fails if unable to parse the shared secret", func() {
				r := scaffold.api.Preregister(scaffold.runtime)
				g.Assert(r.Errors[0].Error()).Equal(defs.ErrInvalidDeviceSharedSecret)
			})
		})

		g.Describe("with a valid request body and a valid rsa shared secret", func() {
			g.BeforeEach(func() {
				secretValue := []byte("30820122300d06092a864886f70d01010105000382010f003082010a0282010100d50ceac3406492f2b4dc91322dbdf6374aca85bd40ac1f4cbd8b9da728f7263c9f7e58d2750bfb4a2b33cb245d1acef4fee544a6ebf583d8b5f691451b95a45410009ba3a524635534523a455d363f5e0afacd983532bd56865afda07545d736004776393682d2e3a7893e672ccdc4e62eae1fafd634d95fb468d29a09261e11279140f5bf98d2be817beffb9d398faf6eeb132ea8e5626c935c33e27021ea878463cf998543625af88dacb29679a19fbf977ffb3d80692ff65236ebee3f9b503dc78ba879f7113c7cd1c689b73050266548c37470e6ece176d24ce4312d81de21923dd2e6a4749fc84451972ee02fd12cbaeb265e6eec8bb814fe6a5dac2cdf0203010001")

				body := []byte(fmt.Sprintf(`
				{
					"name": "some-device",
					"shared_secret": "%s"
				}
				`, secretValue))

				scaffold.body.Write(body)
			})

			g.It("errors when unable to allocate a registration with the registry", func() {
				scaffold.registry.allocationErrors = append(scaffold.registry.allocationErrors, fmt.Errorf("error"))
				r := scaffold.api.Preregister(scaffold.runtime)
				g.Assert(r.Errors[0].Error()).Equal(defs.ErrServerError)
			})

			g.It("succeeds when able to allocate a registration with the registry", func() {
				r := scaffold.api.Preregister(scaffold.runtime)
				g.Assert(len(r.Errors)).Equal(0)
			})
		})

	})
}
