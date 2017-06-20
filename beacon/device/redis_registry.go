package device

import "log"
import "github.com/garyburd/redigo/redis"
import "github.com/dadleyy/beacon.api/beacon/defs"

type RedisRegistry struct {
	*log.Logger
	redis.Conn
}

func (registry *RedisRegistry) Remove(uuid string) error {
	_, e := registry.Do("LREM", defs.RedisDeviceListKey, 1, uuid)
	return e
}

func (registry *RedisRegistry) Exists(id string) bool {
	response, e := registry.Do("LRANGE", defs.RedisDeviceListKey, 0, -1)

	if e != nil {
		return false
	}

	strings, e := redis.Strings(response, e)

	if e != nil || len(strings) == 0 {
		return false
	}

	for _, s := range strings {
		if s == id {
			return true
		}
	}

	return false
}

func (registry *RedisRegistry) Insert(id string) error {
	_, e := registry.Do("LPUSH", defs.RedisDeviceListKey, id)
	return e
}
