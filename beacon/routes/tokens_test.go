package routes

import "fmt"
import "bytes"
import "testing"
import "crypto/rand"
import "encoding/hex"
import "net/http/httptest"
import "github.com/franela/goblin"
import "github.com/dadleyy/beacon.api/beacon/net"
import "github.com/dadleyy/beacon.api/beacon/defs"
import "github.com/dadleyy/beacon.api/beacon/device"

type tokensAPIScaffolding struct {
	api     *TokensAPI
	store   *testDeviceTokenStore
	index   *testDeviceIndex
	runtime *net.RequestRuntime
	body    *bytes.Buffer
}

func (t *tokensAPIScaffolding) Reset() {
	logger := newTestRouteLogger()

	t.store = &testDeviceTokenStore{}
	t.index = &testDeviceIndex{}

	t.body = bytes.NewBuffer([]byte{})

	t.runtime = &net.RequestRuntime{
		Request: httptest.NewRequest("GET", "/tokens", t.body),
	}

	t.api = &TokensAPI{
		LeveledLogger: logger,
		TokenStore:    t.store,
		Index:         t.index,
	}
}

func Test_TokensAPI(suite *testing.T) {
	g := goblin.Goblin(suite)

	scaffold := &tokensAPIScaffolding{}

	g.Describe("CreateToken", func() {

		g.BeforeEach(scaffold.Reset)

		g.It("fails with an invalid request body", func() {
			r := scaffold.api.CreateToken(scaffold.runtime)
			g.Assert(r.Errors[0].Error()).Equal(defs.ErrInvalidTokenRequest)
		})

		g.It("fails if the request's name value is invalid", func() {
			scaffold.body.Write([]byte(`{}`))
			r := scaffold.api.CreateToken(scaffold.runtime)
			g.Assert(r.Errors[0].Error()).Equal(defs.ErrInvalidDeviceTokenName)
		})

		g.Describe("with a valid name field", func() {

			g.BeforeEach(func() {
				nameBuffer := make([]byte, defs.SecurityUserDeviceNameMinLength+1)
				rand.Read(nameBuffer)
				json := fmt.Sprintf(`{"name": "%s"}`, hex.EncodeToString(nameBuffer))
				scaffold.body.Write([]byte(json))
			})

			g.It("fails if it is unable to find the device associated with the request", func() {
				scaffold.index.findErrors = append(scaffold.index.findErrors, fmt.Errorf("bad-find"))
				r := scaffold.api.CreateToken(scaffold.runtime)
				g.Assert(r.Errors[0].Error()).Equal(defs.ErrNotFound)
			})

			g.It("fails if no token was provided in the header", func() {
				scaffold.index.foundDevices = append(scaffold.index.foundDevices, device.RegistrationDetails{})
				r := scaffold.api.CreateToken(scaffold.runtime)
				g.Assert(r.Errors[0].Error()).Equal(defs.ErrInvalidTokenRequest)
			})

			g.Describe("with a valid name and device id", func() {

				deviceID := "some-device"

				g.BeforeEach(func() {
					nameBuffer := make([]byte, defs.SecurityUserDeviceNameMinLength+1)
					rand.Read(nameBuffer)
					json := fmt.Sprintf(`{"name": "%s", "device_id": "%s"}`, hex.EncodeToString(nameBuffer), deviceID)
					scaffold.body.Reset()
					scaffold.body.Write([]byte(json))
					scaffold.runtime.Header.Set(defs.APIUserTokenHeader, "some-token")
					scaffold.index.foundDevices = append(scaffold.index.foundDevices, device.RegistrationDetails{
						DeviceID: deviceID,
					})
				})

				g.It("fails if it is unable to authorize the token found in the header", func() {
					r := scaffold.api.CreateToken(scaffold.runtime)
					g.Assert(r.Errors[0].Error()).Equal(defs.ErrInvalidTokenRequest)
					v, ok := scaffold.store.authorizationAttempts[deviceID]
					g.Assert(ok).Equal(true)
					permission, ok := v["some-token"]
					g.Assert(ok).Equal(true)
					g.Assert(permission).Equal(uint(defs.SecurityDeviceTokenPermissionAdmin))
				})

				g.It("errors if it is unable to create the token", func() {
					scaffold.store.authorized = true
					r := scaffold.api.CreateToken(scaffold.runtime)
					g.Assert(r.Errors[0].Error()).Equal(defs.ErrServerError)
				})

				g.It("succeeds if it is unable to create the token", func() {
					scaffold.store.authorized = true
					scaffold.store.createdTokens = append(scaffold.store.createdTokens, device.TokenDetails{})
					r := scaffold.api.CreateToken(scaffold.runtime)
					g.Assert(len(r.Errors)).Equal(0)
				})
			})

		})

	})
}
