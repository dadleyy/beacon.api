package device

import "github.com/dadleyy/beacon.api/beacon/interchange"

// FeedbackStore defines an interface that logs device state into a persisted store.
type FeedbackStore interface {
	LogFeedback(interchange.FeedbackMessage) error
	ListFeedback(string, int) ([]interchange.FeedbackMessage, error)
}
