package device

import "log"
import "fmt"
import "bytes"
import "testing"
import "github.com/rafaeljusto/redigomock"
import "github.com/dadleyy/beacon.api/beacon/defs"
import "github.com/dadleyy/beacon.api/beacon/logging"

func Test_FindDevice(describe *testing.T) {
	out := bytes.NewBuffer([]byte{})
	logger := log.New(out, "", 0)
	logger.SetFlags(0)
	mock := redigomock.NewConn()

	r := RedisRegistry{
		Logger: &logging.Logger{logger},
		Conn:   mock,
	}

	describe.Run("with no devices in the store", func(it *testing.T) {
		defer mock.Clear()
		if _, e := r.FindDevice("garbage"); e == nil {
			it.Fail()
		}
	})

	describe.Run("with a device in the store", func(describe *testing.T) {
		deviceID, deviceName, deviceSecret := "123", "some-device", "9876"
		registryKey := r.genRegistryKey(deviceID)

		describe.Run("but unable to load data", func(it *testing.T) {
			defer mock.Clear()
			mock.Command("EXISTS", registryKey).Expect([]byte("true"))
			_, e := r.FindDevice(deviceID)

			if e == nil {
				it.Fail()
			}
		})

		describe.Run("with valid device details & searching by ID", func(it *testing.T) {
			defer mock.Clear()
			mock.Command("EXISTS", registryKey).Expect([]byte("true"))
			mock.Command("HMGET", registryKey, "device:uuid", "device:name", "device:secret").ExpectSlice(
				[]byte(deviceID),
				[]byte(deviceName),
				[]byte(deviceSecret),
			)

			details, e := r.FindDevice(deviceID)

			if e != nil {
				it.Fatalf("expected to find device but received: %s", e.Error())
			}

			if details.SharedSecret != deviceSecret || details.Name != deviceName {
				it.Fatalf("expected to find device but received: %v", details)
			}
		})

		describe.Run("recevied a mismatch during the loading from HMGET", func(it *testing.T) {
			defer mock.Clear()
			mock.Command("EXISTS", r.genRegistryKey(deviceName)).Expect([]byte("false"))
			mock.Command("KEYS", fmt.Sprintf("%s*", defs.RedisDeviceRegistryKey)).ExpectSlice(
				[]byte(r.genRegistryKey(deviceID)),
			)
			mock.Command("HMGET", r.genRegistryKey(deviceID), "device:name", "device:uuid", "device:secret").ExpectSlice(
				[]byte("not-the-same"),
				[]byte("not-the-same"),
				[]byte("not-the-same"),
			)

			details, e := r.FindDevice(deviceName)

			if e == nil {
				it.Fatalf("expected to be unable to load device but received: %v", details)
			}
		})

		describe.Run("with valid device details & searching by name", func(it *testing.T) {
			defer mock.Clear()
			mock.Command("EXISTS", r.genRegistryKey(deviceName)).Expect([]byte("false"))
			mock.Command("KEYS", fmt.Sprintf("%s*", defs.RedisDeviceRegistryKey)).ExpectSlice(
				[]byte(r.genRegistryKey(deviceID)),
			)
			mock.Command("HMGET", r.genRegistryKey(deviceID), "device:name", "device:uuid", "device:secret").ExpectSlice(
				[]byte(deviceName),
				[]byte(deviceID),
				[]byte(deviceSecret),
			)

			details, e := r.FindDevice(deviceName)

			if e != nil {
				it.Fatalf("expected to find device but received: %s", e.Error())
			}

			if details.SharedSecret != deviceSecret || details.Name != deviceName {
				it.Fatalf("expected to find device but received: %v", details)
			}
		})
	})
}
