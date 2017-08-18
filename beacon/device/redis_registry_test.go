package device

import "os"
import "log"
import "testing"
import "github.com/rafaeljusto/redigomock"
import "github.com/dadleyy/beacon.api/beacon/logging"

func Test_FindDevice(describe *testing.T) {
	logger := log.New(os.Stdout, "", 0)
	logger.SetFlags(0)
	mock := redigomock.NewConn()

	r := RedisRegistry{
		Logger: &logging.Logger{logger},
		Conn:   mock,
	}

	describe.Run("with no devices in the store", func(it *testing.T) {
		mock.Clear()
		if _, e := r.FindDevice("garbage"); e == nil {
			it.Fail()
		}
	})

	describe.Run("with a device in the store", func(describe *testing.T) {
		deviceID := "123"
		registryKey := r.genRegistryKey(deviceID)

		setup := func() {
			mock.Clear()
			mock.Command("EXISTS", registryKey).Expect([]byte("true"))
		}

		describe.Run("but unable to load data", func(it *testing.T) {
			setup()
			_, e := r.FindDevice(deviceID)

			if e == nil {
				it.Fail()
			}
		})

		describe.Run("with valid device details & searching by ID", func(it *testing.T) {
			setup()
			mock.Command("HMGET", registryKey, "device:uuid", "device:name", "device:secret").ExpectSlice(
				[]byte(deviceID),
				[]byte("someDevice"),
				[]byte("omg-secret"),
			)

			details, e := r.FindDevice(deviceID)

			if e != nil {
				it.Fail()
			}

			if details.SharedSecret != "omg-secret" || details.Name != "someDevice" {
				it.Fail()
			}
		})
	})
}
