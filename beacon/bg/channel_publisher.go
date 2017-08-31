package bg

import "io"

// ChannelStore defines a map of channel names to the channel that will send/receivers readers
type ChannelStore map[string]chan io.Reader

// ChannelPublisher defines an interface that sends an io.Reader interface to a consumer
type ChannelPublisher interface {
	PublishReader(string, io.Reader) error
}
