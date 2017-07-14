package defs

const (
	// RedisDeviceIndexKey is the key used by the regis device registry to store device ids
	RedisDeviceIndexKey = "beacon:device_index"

	// RedisDeviceRegistryKey is the key used by the regis device registry to store device information
	RedisDeviceRegistryKey = "beacon:device_registry"

	// RedisDeviceFeedbackKey is the key used by the regis device registry to store device feedback
	RedisDeviceFeedbackKey = "beacon:device_feedback"

	// RedisRegistrationRequestListKey is the key used for registration requests
	RedisRegistrationRequestListKey = "beacon:registration_requests"

	// RedisDeviceIDField is the field that contains the unique id of the device
	RedisDeviceIDField = "device:uuid"

	// RedisDeviceNameField is the field that contains the unique name of the device
	RedisDeviceNameField = "device:name"

	// RedisDeviceSecretField is the field that contains the unique secret of the device
	RedisDeviceSecretField = "device:secret"

	// RedisRegistrationNameField is the redis key used to store registration names
	RedisRegistrationNameField = "registration:name"

	// RedisRegistrationSecretField is the redis key used to store registration secrets
	RedisRegistrationSecretField = "registration:secret"

	// RedisMaxFeedbackEntries is the maximum amount of entries a device is allowed to have at any given time.
	RedisMaxFeedbackEntries = 100
)
