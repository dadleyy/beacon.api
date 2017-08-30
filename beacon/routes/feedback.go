package routes

import "strconv"
import "io/ioutil"
import "github.com/golang/protobuf/proto"

import "github.com/dadleyy/beacon.api/beacon/net"
import "github.com/dadleyy/beacon.api/beacon/defs"
import "github.com/dadleyy/beacon.api/beacon/device"
import "github.com/dadleyy/beacon.api/beacon/logging"
import "github.com/dadleyy/beacon.api/beacon/interchange"

// NewFeedbackAPI returns a new initialized feed back api
func NewFeedbackAPI(store device.FeedbackStore, index device.Index) *Feedback {
	logger := logging.New(defs.FeedbackAPILogPrefix, logging.Green)

	return &Feedback{
		LeveledLogger: logger,
		FeedbackStore: store,
		Index:         index,
	}
}

// Feedback is the route group that handles creating device feedback entries.
type Feedback struct {
	logging.LeveledLogger
	device.FeedbackStore
	device.Index
}

type reportEntry struct {
	Red   uint32 `json:"red"`
	Green uint32 `json:"green"`
	Blue  uint32 `json:"blue"`
}

// ListFeedback validates a payload from the client and adds an entry to the device feedback log.
func (feedback *Feedback) ListFeedback(runtime *net.RequestRuntime) net.HandlerResult {
	count, e := strconv.Atoi(runtime.GetQueryParam("count"))

	if e != nil || count >= 1 != true || count >= 100 {
		count = 1
		feedback.Debugf("defaulting feedback count to 1")
	}

	deviceID := runtime.GetQueryParam("device_id")

	if _, e := feedback.FindDevice(deviceID); e != nil {
		feedback.Warnf("invalid device id: %s", deviceID)
		return runtime.LogicError(defs.ErrNotFound)
	}

	entries, e := feedback.FeedbackStore.ListFeedback(deviceID, count-1)

	if e != nil {
		feedback.Warnf("unable to load device feedback: %s", e.Error())
		return runtime.ServerError()
	}

	feedback.Debugf("found %d entries for device %s", len(entries), runtime.GetQueryParam("device_id"))

	results := make([]interface{}, 0, len(entries))

	for _, top := range entries {
		payload := top.GetPayload()

		if payload == nil || len(payload) == 0 {
			results = append(results, nil)
			continue
		}

		switch top.Type {
		case interchange.FeedbackMessageType_ERROR:
			results = append(results, nil)
		case interchange.FeedbackMessageType_REPORT:
			report := interchange.ReportMessage{}

			if e := proto.Unmarshal(payload, &report); e != nil {
				feedback.Errorf("unable to unmarshal latest feedback payload: %s", e.Error())
				return runtime.LogicError(defs.ErrBadInterchangeData)
			}

			results = append(results, reportEntry{report.Red, report.Green, report.Blue})
		}
	}

	return net.HandlerResult{Results: results}
}

// CreateFeedback validates a payload from the client and adds an entry to the device feedback log.
func (feedback *Feedback) CreateFeedback(runtime *net.RequestRuntime) net.HandlerResult {
	buf, e := ioutil.ReadAll(runtime.Body)

	if e != nil {
		runtime.Errorf("invalid data recieved in feedback api: %s", e.Error())
		return runtime.LogicError("invalid-request")
	}

	if runtime.ContentType() != defs.APIFeedbackContentTypeHeader {
		runtime.Warnf("invalid content type for feedback: %s", runtime.ContentType())
		return runtime.LogicError("invalid-content-type")
	}

	message := interchange.FeedbackMessage{}

	if e := proto.Unmarshal(buf, &message); e != nil {
		runtime.Errorf("invalid data recieved in feedback api: %s", e.Error())
		return runtime.LogicError("invalid-request")
	}

	auth := message.GetAuthentication()

	if auth == nil {
		runtime.Errorf("unable to load authentication from message")
		return runtime.LogicError("invalid-request")
	}

	if e := feedback.LogFeedback(message); e != nil {
		runtime.Warnf("unable to find device: %s", e.Error())
		return runtime.LogicError("not-found")
	}

	runtime.Infof("successfully posted feedback from device[%s]", auth.DeviceID)
	return net.HandlerResult{}
}
