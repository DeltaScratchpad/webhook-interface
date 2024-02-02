package webhook_tracker

type WebhookState interface {

	// IncrementCallCount Increment the call count for a given webhook and query ID.
	// Returns the new call count
	IncrementCallCount(webhook string, queryID string) int64

	// HasBeenCalled Check if a webhook has been called for a given query ID
	HasBeenCalled(webhook string, queryID string) bool
}
