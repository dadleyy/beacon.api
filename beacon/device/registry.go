package device

// RegistrationRequest holds the information for a pending registration
type RegistrationRequest struct {
	SharedSecret string `json:"-"`
	Name         string `json:"name"`
}

// RegistrationDetails holds the information about a given device connection
type RegistrationDetails struct {
	SharedSecret string `json:"-"`
	Name         string `json:"name"`
	DeviceID     string `json:"device_id"`
}

// Registry is an interface for allocating and filling registration requests
type Registry interface {
	Index
	ListRegistrations() ([]RegistrationDetails, error)
	FillRegistration(string, string) error
	AllocateRegistration(RegistrationRequest) error
}
