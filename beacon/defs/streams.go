package defs

import "io"

// BackgroundChannels defines a map of channel names to the channel that will send/receivers readers
type BackgroundChannels map[string]chan io.Reader

const (
	// DeviceControlChannelName is the name of the stream that will be used to send device control messages along
	DeviceControlChannelName = "chan:device-control"

	// DeviceFeedbackChannelName is the name of the stream that will broacast messages received from devices
	DeviceFeedbackChannelName = "chan:device-feedback"
)
