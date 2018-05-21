package bg

import "io"
import "fmt"
import "github.com/dadleyy/beacon.api/beacon/defs"

// ChannelPublisher defines an interface that sends an io.Reader interface to a consumer
type ChannelPublisher interface {
	PublishReader(string, io.Reader) error
}

// ChannelStore defines a map of channel names to the channel that will send/receivers readers
type ChannelStore map[string]chan io.Reader

// PublishReader publishes an instance of an io.Reader to a channel it owns.
func (s *ChannelStore) PublishReader(name string, reader io.Reader) error {
	if s == nil {
		return fmt.Errorf("invalid-store")
	}

	c, e := (*s)[name]

	if e != true {
		return fmt.Errorf(defs.ErrInvalidBackgroundChannel)
	}

	c <- reader

	return nil
}
