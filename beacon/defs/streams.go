package defs

import "io"

type BackgroundChannels map[string]chan io.Reader

const DeviceControlChannelName = "chan:device-control"
const DeviceFeedbackChannelName = "chan:device-feedback"
