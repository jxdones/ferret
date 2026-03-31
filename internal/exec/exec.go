package exec

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jxdones/ferret/internal/collection"
	"github.com/jxdones/ferret/internal/env"
)

// Default headers used when not specified in the request.
const (
	defaultAcceptHeader = "*/*"
	defaultUserAgent    = "ferret"
	maxResponseBodySize = 10 * 1024 * 1024 // 10MB
	defaultTimeout      = 30 * time.Second
)

// Result represents the result of an HTTP request.
type Result struct {
	Status         int
	StatusText     string
	Proto          string
	Headers        map[string][]string
	Body           []byte
	Duration       time.Duration
	Size           int
	URL            string
	Trace          Trace
	ResponseTooBig bool  // true when response exceeded the 10MB threshold
	ResponseSize   int64 // content-length header, or measured size if not present
}

// Trace represents the trace of an HTTP request.
type Trace struct {
	Events     []TraceEvent
	Redirects  []string
	Proto      string
	RemoteAddr string
}

// TraceEvent represents an event in the trace of an HTTP request.
type TraceEvent struct {
	Name    string
	Elapsed time.Duration
}

// Execute sends an HTTP request and returns the result.
func Execute(ctx context.Context, req collection.Request, e *env.Env) (Result, error) {
	// If no deadline is set, set a default timeout.
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}

	httpReq, err := buildHTTPRequest(ctx, req, e)
	if err != nil {
		return Result{}, err
	}

	rec := newTraceRecorder()
	httpReq = httpReq.WithContext(rec.contextWithTrace(ctx))
	client := httpClientWithRedirects(rec)

	resp, err := client.Do(httpReq)
	if err != nil {
		return Result{}, fmt.Errorf("exec: send request: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(rec.start)
	rec.record("response headers received")

	knownSize := resp.ContentLength                        // -1 if unknown
	lr := io.LimitReader(resp.Body, maxResponseBodySize+1) // +1 to detect truncation
	respBody, err := io.ReadAll(lr)
	if err != nil {
		return Result{}, fmt.Errorf("exec: read response body: %w", err)
	}

	responseTooBig := len(respBody) > maxResponseBodySize
	var responseSize int64
	if responseTooBig {
		respBody = nil
		if knownSize > 0 {
			responseSize = knownSize
		} else {
			responseSize = maxResponseBodySize
		}
	} else {
		responseSize = int64(len(respBody))
	}

	rec.record("response body read")
	if resp.TLS != nil {
		rec.trace.Proto = resp.TLS.NegotiatedProtocol
	}
	return resultFromResponse(resp, respBody, duration, *rec.trace, responseTooBig, responseSize), nil
}

// resultFromResponse creates a Result from an *http.Response and the trace.
func resultFromResponse(resp *http.Response, body []byte, duration time.Duration, trace Trace, responseTooBig bool, responseSize int64) Result {
	return Result{
		Status:         resp.StatusCode,
		StatusText:     resp.Status,
		Proto:          resp.Proto,
		Headers:        map[string][]string(resp.Header),
		Body:           body,
		Duration:       duration,
		Size:           len(body),
		URL:            resp.Request.URL.String(),
		Trace:          trace,
		ResponseTooBig: responseTooBig,
		ResponseSize:   responseSize,
	}
}

// buildHTTPRequest applies env interpolation, validates the URL, and builds an
// *http.Request.
func buildHTTPRequest(ctx context.Context, req collection.Request, e *env.Env) (*http.Request, error) {
	url := interpolate(req.URL, e)
	if strings.TrimSpace(url) == "" {
		return nil, fmt.Errorf("url is required")
	}
	if !strings.Contains(url, "://") {
		url = "https://" + url
	}
	body := interpolate(req.Body, e)
	if v := unresolvedVars(url); v != "" {
		return nil, fmt.Errorf("unresolved variable %s - add it to environments/*.yaml and load that env (e.g. TUI or ferret run -e <name>), or export it in the shell", v)
	}

	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("exec: build request: %w", err)
	}
	for k, v := range req.Headers {
		httpReq.Header.Set(k, interpolate(v, e))
	}
	setDefaultHeaders(httpReq)
	return httpReq, nil
}

// setDefaultHeaders sets default headers if not already set.
func setDefaultHeaders(req *http.Request) {
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", defaultAcceptHeader)
	}
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", defaultUserAgent)
	}
}

// unresolvedVars returns the first {{key}} placeholder in a string, or an empty
// string if no placeholder is found.
func unresolvedVars(s string) string {
	start := strings.Index(s, "{{")
	if start == -1 {
		return ""
	}
	end := strings.Index(s[start:], "}}")
	if end == -1 {
		return ""
	}
	return s[start : start+end+2]
}

// interpolate replaces all {{key}} placeholders in a string with values
// resolved from the environment.
func interpolate(s string, e *env.Env) string {
	for _, m := range []map[string]string{e.Shell, e.Session, e.File} {
		for k, v := range m {
			s = strings.ReplaceAll(s, "{{"+k+"}}", v)
		}
	}
	return s
}
