package webhook_tracker

import (
	"sync"
	"sync/atomic"
)

type LocalWebhookState struct {
	callCounts map[string]*atomic.Int64
	callLock   *sync.RWMutex
}

func NewLocalWebhookState() *LocalWebhookState {
	return &LocalWebhookState{
		callCounts: make(map[string]*atomic.Int64),
		callLock:   new(sync.RWMutex),
	}
}

func (l *LocalWebhookState) IncrementCallCount(webhook string, queryID string) int64 {

	// First optimistically just try to use a read lock
	l.callLock.RLock()
	val, ok := l.callCounts[webhook+queryID]
	if ok {
		l.callLock.RUnlock()
		return val.Add(1)
	} else {
		l.callLock.RUnlock()
		l.callLock.Lock()
		val, ok = l.callCounts[webhook+queryID]
		if ok {
			l.callLock.Unlock()
			return val.Add(1)
		} else {
			var counter atomic.Int64
			counter.Store(1)
			l.callCounts[webhook+queryID] = &counter
			l.callLock.Unlock()
			return 1
		}
	}
}

func (l *LocalWebhookState) HasBeenCalled(webhook string, queryID string) bool {
	l.callLock.RLock()
	defer l.callLock.RUnlock()
	_, ok := l.callCounts[webhook+queryID]
	return ok
}
