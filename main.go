package requestid

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
)

const (
	requestIdHeader     = "X-Request-Id"
	correlationIdHeader = "X-Correlation-Id"
)

type Config struct {
	RequestIdHeader     string `json:"requestIdHeader,omitempty"`
	CorrelationIdHeader string `json:"correlationIdHeader,omitempty"`
}

func CreateConfig() *Config {
	return &Config{
		RequestIdHeader:     requestIdHeader,
		CorrelationIdHeader: correlationIdHeader,
	}
}

type RequestID struct {
	next                http.Handler
	requestIdHeader     string
	correlationIdHeader string
}

func New(_ context.Context, next http.Handler, cfg *Config, _ string) (http.Handler, error) {
	if cfg.RequestIdHeader == "" {
		cfg.RequestIdHeader = requestIdHeader
	}
	if cfg.CorrelationIdHeader == "" {
		cfg.CorrelationIdHeader = correlationIdHeader
	}
	return &RequestID{
		next:                next,
		requestIdHeader:     cfg.RequestIdHeader,
		correlationIdHeader: cfg.CorrelationIdHeader,
	}, nil
}

func (m *RequestID) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	id := newID()
	req.Header.Set(m.requestIdHeader, id)
	rw.Header().Set(m.requestIdHeader, id)

	correlationID := req.Header.Get(m.correlationIdHeader)
	if !isValidUUIDv4(correlationID) {
		correlationID = id
		req.Header.Set(m.correlationIdHeader, correlationID)
	}
	rw.Header().Set(m.correlationIdHeader, correlationID)

	m.next.ServeHTTP(rw, req)
}

func isValidUUIDv4(s string) bool {
	if len(s) != 32 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	if s[12] != '4' {
		return false
	}
	v := s[16]
	return v == '8' || v == '9' || v == 'a' || v == 'b' || v == 'A' || v == 'B'
}

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits
	return fmt.Sprintf("%x", b)
}
