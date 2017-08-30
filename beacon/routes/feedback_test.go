package routes

import "fmt"
import "bytes"
import "testing"
import "net/http/httptest"
import "github.com/franela/goblin"
import "github.com/golang/protobuf/proto"

import "github.com/dadleyy/beacon.api/beacon/net"
import "github.com/dadleyy/beacon.api/beacon/defs"
import "github.com/dadleyy/beacon.api/beacon/device"
import "github.com/dadleyy/beacon.api/beacon/interchange"

type testFeedbackAPIScaffolding struct {
	index   *testDeviceIndex
	store   *testFeedbackStore
	api     *Feedback
	runtime *net.RequestRuntime
	body    *bytes.Buffer
}

func prepareFeedbackAPIScaffold() testFeedbackAPIScaffolding {
	store := testFeedbackStore{}
	index := testDeviceIndex{}

	api := Feedback{
		LeveledLogger: newTestRouteLogger(),
		FeedbackStore: &store,
		Index:         &index,
	}

	body := bytes.NewBuffer([]byte{})

	runtime := net.RequestRuntime{
		Request: httptest.NewRequest("GET", "/feedback", body),
	}

	return testFeedbackAPIScaffolding{
		index:   &index,
		store:   &store,
		api:     &api,
		runtime: &runtime,
		body:    body,
	}
}

func Test_FeedbackAPI(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("ListFeedback", func() {
		var scaffold testFeedbackAPIScaffolding

		g.BeforeEach(func() {
			scaffold = prepareFeedbackAPIScaffold()
		})

		g.It("returns an error if unable to find the device", func() {
			scaffold.index.findErrors = append(scaffold.index.findErrors, fmt.Errorf("bad-find"))
			r := scaffold.api.ListFeedback(scaffold.runtime)
			g.Assert(r.Errors[0].Error()).Equal(defs.ErrNotFound)
		})

		g.Describe("having found the device", func() {
			g.BeforeEach(func() {
				scaffold.index.foundDevices = append(scaffold.index.foundDevices, device.RegistrationDetails{})
			})

			g.It("fails if unable to list the feedback from the store", func() {
				scaffold.store.listErrors = append(scaffold.store.listErrors, fmt.Errorf("bad-list"))
				r := scaffold.api.ListFeedback(scaffold.runtime)
				g.Assert(r.Errors[0].Error()).Equal(defs.ErrServerError)
			})

			g.It("returns nil for feedback items without a payload", func() {
				scaffold.store.listResults = append(scaffold.store.listResults, interchange.FeedbackMessage{})
				r := scaffold.api.ListFeedback(scaffold.runtime)
				list, _ := r.Results.([]interface{})
				first, _ := list[0].(error)
				g.Assert(first).Equal(nil)
			})

			g.It("returns nil for error reporting feedback items", func() {
				scaffold.store.listResults = append(scaffold.store.listResults, interchange.FeedbackMessage{
					Type:    interchange.FeedbackMessageType_ERROR,
					Payload: []byte("empty"),
				})
				r := scaffold.api.ListFeedback(scaffold.runtime)
				list, _ := r.Results.([]interface{})
				first, _ := list[0].(error)
				g.Assert(first).Equal(nil)
			})

			g.It("returns an an error when unable to unmarshall the payload of a report entry", func() {
				payload := []byte("this-is-ugly")

				scaffold.store.listResults = append(scaffold.store.listResults, interchange.FeedbackMessage{
					Type:    interchange.FeedbackMessageType_REPORT,
					Payload: payload,
				})

				r := scaffold.api.ListFeedback(scaffold.runtime)
				g.Assert(r.Errors[0].Error()).Equal(defs.ErrBadInterchangeData)
			})

			g.It("returns an unmarshalled report entry for proper items", func() {
				payload, _ := proto.Marshal(&interchange.ReportMessage{
					Red:   100,
					Green: 200,
					Blue:  300,
				})

				scaffold.store.listResults = append(scaffold.store.listResults, interchange.FeedbackMessage{
					Type:    interchange.FeedbackMessageType_REPORT,
					Payload: payload,
				})

				r := scaffold.api.ListFeedback(scaffold.runtime)
				list, _ := r.Results.([]interface{})
				first, ok := list[0].(reportEntry)

				g.Assert(ok).Equal(true)
				g.Assert(first.Red).Equal(uint(100))
				g.Assert(first.Green).Equal(uint(200))
				g.Assert(first.Blue).Equal(uint(300))
			})
		})
	})

	g.Describe("CreateFeedback", func() {
		var scaffold testFeedbackAPIScaffolding

		g.BeforeEach(func() {
			scaffold = prepareFeedbackAPIScaffold()
		})

		g.It("returns an error without a proper content type header", func() {
			r := scaffold.api.CreateFeedback(scaffold.runtime)
			g.Assert(r.Errors[0].Error()).Equal(defs.ErrInvalidContentType)
		})

		g.It("returns an error with an invalid body", func() {
			scaffold.runtime.Header.Set(defs.APIContentTypeHeader, defs.APIFeedbackContentTypeHeader)
			r := scaffold.api.CreateFeedback(scaffold.runtime)
			g.Assert(r.Errors[0].Error()).Equal(defs.ErrBadInterchangeData)
		})
	})
}
