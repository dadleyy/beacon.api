package device

import "log"
import "fmt"
import "bytes"
import "testing"
import "github.com/rafaeljusto/redigomock"
import "github.com/dadleyy/beacon.api/beacon/defs"
import "github.com/dadleyy/beacon.api/beacon/logging"

const (
	permissionField = defs.RedisDeviceTokenPermissionField
)

func subject() (RedisRegistry, *redigomock.Conn) {
	out := bytes.NewBuffer([]byte{})
	logger := log.New(out, "", 0)
	logger.SetFlags(0)
	mock := redigomock.NewConn()

	return RedisRegistry{
		Logger: &logging.Logger{Logger: logger},
		Conn:   mock,
	}, mock
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

		describe.Run("recevied an error during the loading from KEYS", func(it *testing.T) {
			defer mock.Clear()
			mock.Command("EXISTS", r.genRegistryKey(deviceName)).Expect([]byte("false"))
			mock.Command("KEYS", fmt.Sprintf("%s*", defs.RedisDeviceRegistryKey)).ExpectError(fmt.Errorf("problems"))

			details, e := r.FindDevice(deviceName)

			if e == nil {
				it.Fatalf("expected to be unable to load device but received: %v", details)
			}
		})

		describe.Run("recevied an error during the parsing of strings from KEYS", func(it *testing.T) {
			defer mock.Clear()
			mock.Command("EXISTS", r.genRegistryKey(deviceName)).Expect([]byte("false"))
			mock.Command("KEYS", fmt.Sprintf("%s*", defs.RedisDeviceRegistryKey)).Expect(nil)

			details, e := r.FindDevice(deviceName)

			if e == nil {
				it.Fatalf("expected to be unable to load device but received: %v", details)
			}
		})

		describe.Run("recevied an error during the loading of keys via second HMGET", func(it *testing.T) {
			defer mock.Clear()
			mock.Command("EXISTS", r.genRegistryKey(deviceName)).Expect([]byte("false"))
			mock.Command("KEYS", fmt.Sprintf("%s*", defs.RedisDeviceRegistryKey)).ExpectSlice(
				[]byte(r.genRegistryKey(deviceID)),
			)

			mock.Command(
				"HMGET",
				r.genRegistryKey(deviceID), "device:name", "device:uuid", "device:secret",
			).ExpectError(fmt.Errorf("problem"))

			details, e := r.FindDevice(deviceName)

			if e == nil {
				it.Fatalf("expected to be unable to load device but received: %v", details)
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

	describe.Run("with an invalid registration config (no secret)", func(it *testing.T) {
		defer mock.Clear()
		deviceName, deviceSecret := "aaaaaaaaaaaaaaaaaaaa", ""
		e := r.AllocateRegistration(RegistrationRequest{
			Name:         deviceName,
			SharedSecret: deviceSecret,
		})

		if e == nil {
			it.Fatalf("expected error but received nil")
		}
	})

	describe.Run("with an invalid registration config (no name)", func(it *testing.T) {
		defer mock.Clear()
		deviceName, deviceSecret := "", "iiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiii"
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
		mock.Command("HMSET")

		e := r.AllocateRegistration(RegistrationRequest{
			Name:         deviceName,
			SharedSecret: deviceSecret,
		})

		if e != nil {
			it.Fatalf("expected no error but received: %v", e)
		}
	})

	describe.Run("when receiving an error during hset", func(it *testing.T) {
		defer mock.Clear()
		mock.Command("HSET").ExpectError(fmt.Errorf("e"))
		deviceName, deviceSecret := "another-device", "iiiiiiiiiiiiiiiiiiii"

		e := r.AllocateRegistration(RegistrationRequest{
			Name:         deviceName,
			SharedSecret: deviceSecret,
		})

		if e == nil {
			it.Fatalf("expected no error but received: %v", e)
		}
	})

	describe.Run("when receiving an error during hset", func(it *testing.T) {
		defer mock.Clear()
		mock.Command("HSET").ExpectError(fmt.Errorf("e"))
		deviceName, deviceSecret := "another-device", "iiiiiiiiiiiiiiiiiiii"

		e := r.AllocateRegistration(RegistrationRequest{
			Name:         deviceName,
			SharedSecret: deviceSecret,
		})

		if e == nil {
			it.Fatalf("expected no error but received: %v", e)
		}
	})
}

func Test_FillRegistration(describe *testing.T) {
	r, mock := subject()

	describe.Run("when received an error from the initial key lookup", func(it *testing.T) {
		defer mock.Clear()
		e := r.FillRegistration("secret", "uuid")

		if e == nil {
			it.Fatalf("expected error without any device lookup possible")
		}
	})

	describe.Run("when received an invalid response from the initial key lookup", func(it *testing.T) {
		defer mock.Clear()
		mock.Command("KEYS").Expect(nil)
		e := r.FillRegistration("secret", "uuid")

		if e == nil {
			it.Fatalf("expected error without any device lookup possible")
		}
	})

	describe.Run("when received no keys from the initial key lookup", func(it *testing.T) {
		defer mock.Clear()
		mock.Command("KEYS").ExpectSlice([]byte{})
		e := r.FillRegistration("secret", "uuid")

		if e == nil {
			it.Fatalf("expected error without any device lookup possible")
		}
	})

	describe.Run("when received some keys but fails on string conv", func(it *testing.T) {
		defer mock.Clear()
		mock.Command("KEYS").ExpectSlice([]byte("hello"))
		mock.Command("HGET").Expect(nil)
		e := r.FillRegistration("secret", "uuid")

		if e == nil {
			it.Fatalf("expected error without any device lookup possible")
		}
	})

	describe.Run("when received some keys w/o a match from the initial key lookup", func(it *testing.T) {
		defer mock.Clear()
		mock.Command("KEYS").ExpectSlice([]byte("hello"))
		mock.Command("HGET").Expect([]byte("secret"))
		e := r.FillRegistration("secret", "uuid")

		if e == nil {
			it.Fatalf("expected error without any device lookup possible")
		}
	})
}

func TestFindToken(describe *testing.T) {
	r, mock := subject()

	describe.Run("without the ability to find a token", func(it *testing.T) {
		defer mock.Clear()
		d, e := r.FindToken("not-there")

		if e == nil {
			it.Fatalf("expected error but received: %v", d)
		}
	})

	describe.Run("having received an invalid permission mask", func(it *testing.T) {
		defer mock.Clear()
		mock.Command("HGET").Expect([]byte("zzz"))
		d, e := r.FindToken("not-there")

		if e == nil {
			it.Fatalf("expected error but received: %v", d)
		}
	})

	describe.Run("with the ability to find a token", func(it *testing.T) {
		defer mock.Clear()

		token := TokenDetails{
			TokenID:  "123",
			Name:     "some-token",
			DeviceID: "321",
		}

		mock.Command("HGET").Expect([]byte("111"))

		mock.Command("HMGET").ExpectSlice(
			[]byte(token.TokenID),
			[]byte(token.Name),
			[]byte(token.DeviceID),
		)

		d, e := r.FindToken("not-there")

		if e != nil {
			it.Fatalf("expected no error but received: %v", e)
		}

		if d.DeviceID != token.DeviceID || d.Name != token.Name {
			it.Fatalf("expected %v but received: %v", token, d)
		}
	})
}

func Test_AuthorizeToken(describe *testing.T) {
	r, mock := subject()

	describe.Run("without the ability to find a token", func(it *testing.T) {
		defer mock.Clear()
		b := r.AuthorizeToken("device-id", "token", 1)

		if b != false {
			it.Fatalf("expected invalid authorization but received successful attempt")
		}
	})

	describe.Run("having found a valid device and given the shared secret", func(it *testing.T) {
		defer mock.Clear()
		deviceID, deviceName, deviceSecret := "device-id", "device-name", "device-secret"
		registryKey := r.genRegistryKey(deviceID)

		mock.Command("EXISTS", registryKey).Expect([]byte("true"))
		mock.Command("HMGET", registryKey, "device:uuid", "device:name", "device:secret").ExpectSlice(
			[]byte(deviceID),
			[]byte(deviceName),
			[]byte(deviceSecret),
		)

		b := r.AuthorizeToken("device-id", deviceSecret, 1)

		if b != true {
			it.Fatalf("expected valid authorization but received successful attempt")
		}
	})

	describe.Run("having found a valid device and given an invalid token", func(it *testing.T) {
		defer mock.Clear()
		deviceID, deviceName, deviceSecret := "device-id", "device-name", "device-secret"
		registryKey := r.genRegistryKey(deviceID)

		mock.Command("EXISTS", registryKey).Expect([]byte("true"))
		mock.Command("HMGET", registryKey, "device:uuid", "device:name", "device:secret").ExpectSlice(
			[]byte(deviceID),
			[]byte(deviceName),
			[]byte(deviceSecret),
		)

		b := r.AuthorizeToken("device-id", "another-token", 1)

		if b != false {
			it.Fatalf("expected valid authorization but received successful attempt")
		}
	})

	describe.Run("having found a valid device and given a valid token AAA", func(it *testing.T) {
		defer mock.Clear()
		deviceID, deviceName, deviceSecret := "device-id", "device-name", "device-secret"
		registryKey := r.genRegistryKey(deviceID)

		token, tokenValue := TokenDetails{
			TokenID:  "123",
			Name:     "some-token",
			DeviceID: "321",
		}, "asdasdasdad"

		mock.Command("EXISTS", registryKey).Expect([]byte("true"))

		mock.Command("HMGET", registryKey, "device:uuid", "device:name", "device:secret").ExpectSlice(
			[]byte(deviceID),
			[]byte(deviceName),
			[]byte(deviceSecret),
		)

		mock.Command("HGET", r.genTokenRegistrationKey(tokenValue), defs.RedisDeviceTokenPermissionField).Expect([]byte("111"))

		mock.Command(
			"HMGET",
			r.genTokenRegistrationKey(tokenValue),
			defs.RedisDeviceTokenIDField,
			defs.RedisDeviceTokenNameField,
			defs.RedisDeviceTokenDeviceIDField,
		).ExpectSlice(
			[]byte(token.TokenID),
			[]byte(token.Name),
			[]byte(token.DeviceID),
		)

		b := r.AuthorizeToken("device-id", tokenValue, 1)

		if b == false {
			it.Fatalf("expected valid authorization but received successful attempt")
		}
	})
}
