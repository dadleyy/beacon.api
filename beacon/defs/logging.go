package defs

import "log"

const (
	// MainLogPrefix is the log prefix for the main go routine
	MainLogPrefix = "[beacon api] "

	// RegistryLogPrefix is the log prefix for the device registry
	RegistryLogPrefix = "[device registry] "

	// ServerRuntimeLogPrefix is the log prefix for the http server runtime
	ServerRuntimeLogPrefix = "[server runtime] "

	// DeviceControlLogPrefix is the log prefix for the device control processor
	DeviceControlLogPrefix = "[device control] "

	// DeviceFeedbackLogPrefix is the log prefix for the device feeback processor
	DeviceFeedbackLogPrefix = "[device feedback] "

	// DefaultLoggerFlags is the bitmask used to create default logging
	DefaultLoggerFlags = log.Ldate | log.Ltime | log.Lshortfile
)
