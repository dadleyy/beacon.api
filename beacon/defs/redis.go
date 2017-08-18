package defs

const (
	// RedisDeviceIndexKey is the key used by the regis device registry to store device ids
	RedisDeviceIndexKey = "beacon:device-index"

	// RedisDeviceRegistryKey is the key used by the regis device registry to store device information
	RedisDeviceRegistryKey = "beacon:device-registry"

	// RedisDeviceFeedbackKey is the key used by the regis device registry to store device feedback
	RedisDeviceFeedbackKey = "beacon:device-feedback"

	// RedisRegistrationRequestListKey is the key used for registration requests
	RedisRegistrationRequestListKey = "beacon:registration-requests"

	// RedisDeviceIDField is the field that contains the unique id of the device
	RedisDeviceIDField = "device:uuid"

	// RedisDeviceNameField is the field that contains the unique name of the device
	RedisDeviceNameField = "device:name"

	// RedisDeviceTokenListKey is the field that contains the list of tokens associated w/ each device
	RedisDeviceTokenListKey = "device:token-list"

	// RedisDeviceTokenRegistrationKey field for device token information (name)
	RedisDeviceTokenRegistrationKey = "device:token"

	// RedisDeviceTokenNameField is the field that contains the unique name of the token
	RedisDeviceTokenNameField = "device-token:name"

	// RedisDeviceTokenDeviceIDField stores the token's device id
	RedisDeviceTokenDeviceIDField = "device-token:device-id"

	// RedisDeviceTokenIDField stores the unique id of the token
	RedisDeviceTokenIDField = "device-token:uuid"

	// RedisDeviceTokenPermissionField is the field that contains the permission of the token
	RedisDeviceTokenPermissionField = "device-token:permission"

	// RedisDeviceSecretField is the field that contains the unique secret of the device
	RedisDeviceSecretField = "device:secret"

	// RedisRegistrationNameField is the redis key used to store registration names
	RedisRegistrationNameField = "registration:name"

	// RedisRegistrationSecretField is the redis key used to store registration secrets
	RedisRegistrationSecretField = "registration:secret"

	// RedisMaxFeedbackEntries is the maximum amount of entries a device is allowed to have at any given time.
	RedisMaxFeedbackEntries = 100
)
