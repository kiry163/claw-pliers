package imap

import (
	"sync"
	"time"
)

type EmailLog struct {
	From       string    `json:"from"`
	Subject    string    `json:"subject"`
	ReceivedAt time.Time `json:"received_at"`
	Summary    string    `json:"summary"`
}

type EmailContent struct {
	Account string    `json:"account"`
	From    string    `json:"from"`
	To      string    `json:"to"`
	Subject string    `json:"subject"`
	Date    time.Time `json:"date"`
	Body    string    `json:"body"`
	UID     uint32    `json:"uid"`
}

type LogStore struct {
	mu      sync.Mutex
	maxSize int
	entries []EmailLog
}

func NewLogStore(maxSize int) *LogStore {
	if maxSize <= 0 {
		maxSize = 200
	}
	return &LogStore{maxSize: maxSize}
}

func (l *LogStore) Add(entry EmailLog) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.entries = append(l.entries, entry)
	if len(l.entries) > l.maxSize {
		l.entries = l.entries[len(l.entries)-l.maxSize:]
	}
}

func (l *LogStore) List(limit int) []EmailLog {
	l.mu.Lock()
	defer l.mu.Unlock()

	if limit <= 0 || limit > len(l.entries) {
		limit = len(l.entries)
	}

	start := len(l.entries) - limit
	if start < 0 {
		start = 0
	}
	result := make([]EmailLog, limit)
	copy(result, l.entries[start:])
	return result
}

func (l *LogStore) Total() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.entries)
}
