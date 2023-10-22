package sse

import (
	"net/http"

	"github.com/r3labs/sse"
	"github.com/stretchr/testify/mock"
)

// StreamerMock is a mocked implementation of an event streamer.
type StreamerMock struct {
	mock.Mock
}

// NewStreamerMock returns a mocked event streamer.
func NewStreamerMock() *StreamerMock {
	return &StreamerMock{}
}

// Close mock.
func (s *StreamerMock) Close() error {
	args := s.Called()
	return args.Error(0)
}

// TrackPayment mock.
func (s *StreamerMock) TrackPayment(rHash, publicKey string, amount uint64) uint64 {
	args := s.Called(rHash, publicKey, amount)
	return args.Get(0).(uint64)
}

// ServerHTTP mock.
func (s *StreamerMock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Called(w, r)
}

// ServerMock is a mocked implementation of event server.
type ServerMock struct {
	mock.Mock
}

// NewServerMock returns a mocked event server.
func NewServerMock() *ServerMock {
	return &ServerMock{}
}

// Close mock.
func (s *ServerMock) Close() {
	s.Called()
}

// CreateStream mock.
func (s *ServerMock) CreateStream(id string) *sse.Stream {
	args := s.Called(id)
	return args.Get(0).(*sse.Stream)
}

// HTTPHandler mock.
func (s *ServerMock) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	s.Called(w, r)
}

// Publish mock.
func (s *ServerMock) Publish(id string, event *sse.Event) {
	s.Called(id, event)
}
