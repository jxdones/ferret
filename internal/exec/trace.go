package exec

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptrace"
	"time"
)

// traceRecorder captures timing and connection details for an HTTP round-trip.
type traceRecorder struct {
	start time.Time
	trace *Trace
}

// newTraceRecorder creates a new trace recorder and records the request started
// event.
func newTraceRecorder() *traceRecorder {
	t := &Trace{}
	rec := &traceRecorder{start: time.Now(), trace: t}
	rec.record("request started")
	return rec
}

// record records an event in the trace.
func (r *traceRecorder) record(name string) {
	r.trace.Events = append(r.trace.Events, TraceEvent{
		Name:    name,
		Elapsed: time.Since(r.start),
	})
}

// clientTrace returns a *httptrace.ClientTrace that records events in the trace.
func (r *traceRecorder) clientTrace() *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		DNSStart: func(httptrace.DNSStartInfo) {
			r.record("dns started")
		},
		DNSDone: func(httptrace.DNSDoneInfo) {
			r.record("dns done")
		},
		ConnectStart: func(_, _ string) {
			r.record("connect started")
		},
		ConnectDone: func(_, _ string, _ error) {
			r.record("connect done")
		},
		TLSHandshakeStart: func() {
			r.record("tls handshake started")
		},
		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
			r.record("tls handshake done")
		},
		GotConn: func(info httptrace.GotConnInfo) {
			r.record("connection acquired")
			r.trace.RemoteAddr = info.Conn.RemoteAddr().String()
		},
		WroteRequest: func(httptrace.WroteRequestInfo) {
			r.record("request sent")
		},
		GotFirstResponseByte: func() {
			r.record("first response byte")
		},
	}
}

// contextWithTrace returns a new context with the trace recorder's client trace
// attached.
func (r *traceRecorder) contextWithTrace(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return httptrace.WithClientTrace(ctx, r.clientTrace())
}

// redirectCapture returns a function that captures redirects in the trace.
func (r *traceRecorder) redirectCapture() func(*http.Request, []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if req != nil && req.URL != nil {
			r.trace.Redirects = append(r.trace.Redirects, req.URL.String())
		}
		return nil
	}
}

// httpClientWithRedirects returns a *http.Client that captures redirects in
// the trace.
func httpClientWithRedirects(rec *traceRecorder) *http.Client {
	client := *http.DefaultClient
	client.CheckRedirect = rec.redirectCapture()
	return &client
}
