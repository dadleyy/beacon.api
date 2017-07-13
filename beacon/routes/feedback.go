package routes

import "io/ioutil"
import "github.com/golang/protobuf/proto"

import "github.com/dadleyy/beacon.api/beacon/net"
import "github.com/dadleyy/beacon.api/beacon/device"
import "github.com/dadleyy/beacon.api/beacon/interchange"

// Feedback is the route group that handles creating device feedback entries.
type Feedback struct {
	device.Index
}

// Create validates a payload from the client and adds an entry to the device feedback log.
func (feedback *Feedback) Create(runtime *net.RequestRuntime) net.HandlerResult {
	buf, e := ioutil.ReadAll(runtime.Body)

	if e != nil {
		runtime.Errorf("invalid data recieved in feedback api: %s", e.Error())
		return runtime.LogicError("invalid-request")
	}

	message := interchange.FeedbackMessage{}

	if e := proto.Unmarshal(buf, &message); e != nil {
		runtime.Errorf("invalid data recieved in feedback api: %s", e.Error())
		return runtime.LogicError("invalid-request")
	}

	runtime.Infof("received feedback from device: %v", message.Authentication)
	return net.HandlerResult{}
}
