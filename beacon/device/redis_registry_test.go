package device

import "log"
import "fmt"
import "bytes"
import "testing"
import "github.com/rafaeljusto/redigomock"
import "github.com/dadleyy/beacon.api/beacon/defs"
import "github.com/dadleyy/beacon.api/beacon/logging"

func subject() (RedisRegistry, *redigomock.Conn) {
	out := bytes.NewBuffer([]byte{})
	logger := log.New(out, "", 0)
	logger.SetFlags(0)
	mock := redigomock.NewConn()

	return RedisRegistry{&logging.Logger{logger}, mock}, mock
}

func Test_FindDevice(describe *testing.T) {
	r, mock := subject()

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

func Test_AllocateRegistration(describe *testing.T) {
	r, mock := subject()

	describe.Run("when unable to generate a key", func(it *testing.T) {
		defer mock.Clear()
		e := r.AllocateRegistration(RegistrationRequest{})
		if e == nil {
			it.Fatalf("expected error w/o ability to set values but received nil")
		}
	})

	describe.Run("with an invalid registration config", func(it *testing.T) {
		defer mock.Clear()
		deviceName, deviceSecret := "another-device", "9876"
		e := r.AllocateRegistration(RegistrationRequest{
			Name:         deviceName,
			SharedSecret: deviceSecret,
		})

		if e == nil {
			it.Fatalf("expected error but received nil")
		}
	})

	describe.Run("with a valid registration config", func(it *testing.T) {
		defer mock.Clear()
		deviceName, deviceSecret := "another-device", "iiiiiiiiiiiiiiiiiiii"
		mock.Command("HSET")

		e := r.AllocateRegistration(RegistrationRequest{
			Name:         deviceName,
			SharedSecret: deviceSecret,
		})

		if e != nil {
			it.Fatalf("expected no error but received: %v", e)
		}
	})

}
