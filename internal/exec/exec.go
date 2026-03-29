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
)

// Result represents the result of an HTTP request.
type Result struct {
	Status     int
	StatusText string
	Proto      string
	Headers    map[string][]string
	Body       []byte
	Duration   time.Duration
	Size       int
	URL        string
	Trace      Trace
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
func Execute(req collection.Request, e *env.Env) (Result, error) {
	httpReq, err := buildHTTPRequest(req, e)
	if err != nil {
		return Result{}, err
	}

	rec := newTraceRecorder()
	httpReq = httpReq.WithContext(rec.contextWithTrace(context.Background()))
	client := httpClientWithRedirects(rec)

	resp, err := client.Do(httpReq)
	if err != nil {
		return Result{}, fmt.Errorf("exec: send request: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(rec.start)
	rec.record("response headers received")

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{}, fmt.Errorf("exec: read response body: %w", err)
	}
	rec.record("response body read")
	if resp.TLS != nil {
		rec.trace.Proto = resp.TLS.NegotiatedProtocol
	}
	return resultFromResponse(resp, respBody, duration, *rec.trace), nil
}

// resultFromResponse creates a Result from an *http.Response and the trace.
func resultFromResponse(resp *http.Response, body []byte, duration time.Duration, trace Trace) Result {
	return Result{
		Status:     resp.StatusCode,
		StatusText: resp.Status,
		Proto:      resp.Proto,
		Headers:    map[string][]string(resp.Header),
		Body:       body,
		Duration:   duration,
		Size:       len(body),
		URL:        resp.Request.URL.String(),
		Trace:      trace,
	}
}

// buildHTTPRequest applies env interpolation, validates the URL, and builds an
// *http.Request.
func buildHTTPRequest(req collection.Request, e *env.Env) (*http.Request, error) {
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

	httpReq, err := http.NewRequest(req.Method, url, bodyReader)
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
