package device

import "log"
import "fmt"
import "github.com/satori/go.uuid"
import "github.com/garyburd/redigo/redis"
import "github.com/dadleyy/beacon.api/beacon/defs"

// RedisRegistry implements the `Registry` interface w/ a redis backend
type RedisRegistry struct {
	*log.Logger
	redis.Conn
}

// Allocate reserves a spot in the registry to be filled later
func (registry *RedisRegistry) Allocate(details RegistrationRequest) error {
	allocationID := uuid.NewV4().String()
	registryKey := registry.genAllocationKey(allocationID)

	if _, e := registry.Do("HSET", registryKey, defs.RedisRegistrationNameField, details.Name); e != nil {
		return e
	}

	if _, e := registry.Do("HSET", registryKey, defs.RedisRegistrationSecretField, details.SharedSecret); e != nil {
		return e
	}

	return nil
}

// Find searches the registry based on a query string for the first matching device id
func (registry *RedisRegistry) Find(query string) (RegistrationDetails, error) {
	if registry.Exists(query) {
		registry.Printf("found device by id: %s", query)
		return registry.loadDetails(registry.genRegistryKey(query))
	}

	response, e := registry.Do("KEYS", fmt.Sprintf("%s*", defs.RedisDeviceRegistryKey))

	if e != nil {
		return RegistrationDetails{}, e
	}

	registryKeys, e := redis.Strings(response, e)

	if e != nil {
		return RegistrationDetails{}, e
	}

	for _, k := range registryKeys {
		fields, e := registry.hmgetstr(k, defs.RedisDeviceNameField, defs.RedisDeviceIDField, defs.RedisDeviceSecretField)

		if e != nil {
			return RegistrationDetails{}, e
		}

		if fields[0] == query || fields[1] == query {
			d := RegistrationDetails{SharedSecret: fields[2], DeviceID: fields[1], Name: fields[0]}
			registry.Printf("found device by query: %s", query)
			return d, nil
		}
	}

	registry.Printf("did not find matching device: %s", query)
	return RegistrationDetails{}, fmt.Errorf("not-found")
}

// Fill searches the pending registrations and adds the new uuid to the index
func (registry *RedisRegistry) Fill(secret, uuid string) error {
	response, e := registry.Do("KEYS", fmt.Sprintf("%s*", defs.RedisRegistrationRequestListKey))

	if e != nil {
		return e
	}

	requestKeys, e := redis.Strings(response, e)

	if e != nil {
		return e
	}

	for _, k := range requestKeys {
		response, e := registry.Do("HGET", k, defs.RedisRegistrationSecretField)

		if e != nil {
			continue
		}

		s, e := redis.String(response, e)

		if e != nil {
			continue
		}

		if s == secret {
			registry.Printf("found matching secret for device[%s], filling", uuid)
			return registry.fill(k, uuid)
		}
	}

	return fmt.Errorf("not-found")
}

// List prints out a list of all the registered devices
func (registry *RedisRegistry) List() ([]RegistrationDetails, error) {
	response, e := registry.Do("LRANGE", defs.RedisDeviceIndexKey, 0, -1)
	var results []RegistrationDetails

	if e != nil {
		return results, e
	}

	ids, e := redis.Strings(response, e)

	if e != nil {
		return results, e
	}

	for _, k := range ids {
		details, e := registry.loadDetails(registry.genRegistryKey(k))

		if e != nil {
			continue
		}

		results = append(results, details)
	}

	return results, nil
}

// Remove executes the LREM command to the redis connection
func (registry *RedisRegistry) Remove(id string) error {
	if _, e := registry.Do("DEL", registry.genRegistryKey(id)); e != nil {
		return e
	}

	registry.Printf("cleaning up device index...")
	_, e := registry.Do("LREM", defs.RedisDeviceIndexKey, 1, id)

	return e
}

// Exists extracts the full list of device keys and searches for the target id
func (registry *RedisRegistry) Exists(id string) bool {
	keys, e := registry.deviceFieldKeys(id)
	return e == nil && len(keys) >= 1
}

// Insert executes the LPUSH command to the redis connection
func (registry *RedisRegistry) Insert(id string) error {
	if _, e := registry.Do("HSET", registry.genRegistryKey(id), defs.RedisDeviceIDField, id); e != nil {
		return e
	}

	_, e := registry.Do("LPUSH", defs.RedisDeviceIndexKey, id)

	return e
}

func (registry *RedisRegistry) deviceFieldKeys(id string) ([]string, error) {
	response, e := registry.Do("HKEYS", registry.genRegistryKey(id))

	if e != nil {
		return nil, e
	}

	return redis.Strings(response, e)
}

// loadDetails returns the device registration details based on a provided device key
func (registry *RedisRegistry) loadDetails(deviceKey string) (RegistrationDetails, error) {
	f := struct {
		id   string
		name string
		key  string
	}{defs.RedisDeviceIDField, defs.RedisDeviceNameField, defs.RedisDeviceSecretField}
	values, e := registry.hmgetstr(deviceKey, f.id, f.name, f.key)

	if e != nil {
		return RegistrationDetails{}, e
	}

	for _, v := range values {
		if filled := len(v) > 1; !filled {
			return RegistrationDetails{}, fmt.Errorf("invalid-device")
		}
	}

	return RegistrationDetails{
		DeviceID:     values[0],
		Name:         values[1],
		SharedSecret: values[2],
	}, nil
}

// loadRequest loads the registration request associated w/ a given key
func (registry *RedisRegistry) loadRequest(requestKey string) (RegistrationRequest, error) {
	f := struct {
		secret string
		name   string
	}{defs.RedisRegistrationSecretField, defs.RedisRegistrationNameField}
	values, e := registry.hmgetstr(requestKey, f.secret, f.name)

	if e != nil {
		return RegistrationRequest{}, e
	}

	for _, v := range values {
		if filled := len(v) > 1; !filled {
			return RegistrationRequest{}, fmt.Errorf("invalid-request")
		}
	}

	return RegistrationRequest{SharedSecret: values[0], Name: values[1]}, nil
}
func (registry *RedisRegistry) genAllocationKey(id string) string {
	return fmt.Sprintf("%s:%s", defs.RedisRegistrationRequestListKey, id)
}

func (registry *RedisRegistry) genRegistryKey(id string) string {
	return fmt.Sprintf("%s:%s", defs.RedisDeviceRegistryKey, id)
}

// hmgetstr is a wrapper around the redis HMGET command where all fields are expected to be strings
func (registry *RedisRegistry) hmgetstr(key string, fields ...string) ([]string, error) {
	args := []interface{}{key}

	for _, f := range fields {
		args = append(args, f)
	}

	response, e := registry.Do("HMGET", args...)

	if e != nil {
		return nil, e
	}

	list, e := redis.Strings(response, e)

	if e != nil {
		return nil, e
	}

	for i, s := range list {
		if empty := len(s) == 0; empty {
			return nil, fmt.Errorf("invalid-entry[%s:%s]", fields[i], s)
		}
	}

	return list, nil
}

// hgetstr is a wrapper around HGET that casts to a string
func (registry *RedisRegistry) hgetstr(key, field string) (string, error) {
	response, e := registry.Do("HGET", key, field)

	if e != nil {
		return "", e
	}

	return redis.String(response, e)
}

// fill is responsible for loading the information stored during the registration request and creating records in both
// the device registry index as well as the device registry (keys w/ device hash information)
func (registry *RedisRegistry) fill(requestKey, deviceID string) error {
	request, e := registry.loadRequest(requestKey)

	if e != nil {
		return e
	}

	if _, e := registry.Do("LPUSH", defs.RedisDeviceIndexKey, deviceID); e != nil {
		return e
	}

	registryKey := registry.genRegistryKey(deviceID)

	f := struct {
		id   string
		name string
		key  string
	}{defs.RedisDeviceIDField, defs.RedisDeviceNameField, defs.RedisDeviceSecretField}

	_, e = registry.Do("HMSET", registryKey, f.id, deviceID, f.name, request.Name, f.key, request.SharedSecret)

	if e != nil {
		return e
	}

	registry.Printf("filling device registry w/ name[%s] id[%s]", request.Name, deviceID)

	defer registry.Do("DEL", requestKey)

	return nil
}
