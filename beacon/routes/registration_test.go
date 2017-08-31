package routes

import "testing"
import "github.com/franela/goblin"
import "github.com/dadleyy/beacon.api/beacon/device"

type registrationAPIScaffolding struct {
	api      *RegistrationAPI
	registry *testDeviceRegistry
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

	return registrationAPIScaffolding{
		api:      &api,
		registry: &registry,
		stream:   stream,
	}
}

func Test_RegistrationAPI(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("Preregister", func() {
		var scaffold registrationAPIScaffolding

		g.BeforeEach(func() {
			scaffold = prepareRegistrationAPIScaffolding()
		})

	})
}
