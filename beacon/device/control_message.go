package device

import "time"

type ControlMessage struct {
	DeviceId string        `json:"device_id"`
	Red      uint8         `json:"red"`
	Blue     uint8         `json:"blue"`
	Green    uint8         `json:"green"`
	LED      uint8         `json:"led"`
	FadeTime time.Duration `json:"fade_time"`
	Duration time.Duration `json:"duration"`
}
