package routes

import "fmt"
import "github.com/dadleyy/beacon.api/beacon/device"

type testDeviceRegistry struct {
	allocationErrors       []error
	findErrors             []error
	fillErrors             []error
	listRegistrationErrors []error
	removalErrors          []error
	activeRegistrations    []device.RegistrationDetails
}

func (t *testDeviceRegistry) AllocateRegistration(device.RegistrationRequest) error {
	return t.latestError(t.allocationErrors)
}

func (t *testDeviceRegistry) FindDevice(string) (device.RegistrationDetails, error) {
	if e := t.latestError(t.findErrors); e != nil {
		return device.RegistrationDetails{}, e
	}

	if len(t.activeRegistrations) >= 1 {
		return t.activeRegistrations[0], nil
	}

	return device.RegistrationDetails{}, fmt.Errorf("not-found")
}

func (t *testDeviceRegistry) FillRegistration(string, string) error {
	return t.latestError(t.fillErrors)
}

func (t *testDeviceRegistry) RemoveDevice(string) error {
	return t.latestError(t.removalErrors)
}

func (t *testDeviceRegistry) ListRegistrations() ([]device.RegistrationDetails, error) {
	if e := t.latestError(t.listRegistrationErrors); e != nil {
		return nil, e
	}

	return t.activeRegistrations, nil
}

func (t *testDeviceRegistry) latestError(errList []error) error {
	if len(errList) >= 1 {
		return errList[0]
	}
	return nil
}

type testDeviceTokenStore struct {
	authorized     bool
	createdTokens  []device.TokenDetails
	creationErrors []error
	listedTokens   []device.TokenDetails
	listedErrors   []error
}

func (t *testDeviceTokenStore) AuthorizeToken(string, string, uint) bool {
	return t.authorized
}

func (t *testDeviceTokenStore) ListTokens(string) ([]device.TokenDetails, error) {
	if len(t.listedErrors) >= 1 {
		return nil, t.listedErrors[0]
	}

	return t.listedTokens, nil
}

func (t *testDeviceTokenStore) CreateToken(string, string, uint) (device.TokenDetails, error) {
	if len(t.createdTokens) >= 1 {
		return t.createdTokens[0], nil
	}

	if len(t.creationErrors) >= 1 {
		return device.TokenDetails{}, t.creationErrors[0]
	}

	return device.TokenDetails{}, fmt.Errorf("not-found")
}
