package exec

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jxdones/ferret/internal/collection"
	"github.com/jxdones/ferret/internal/env"
)

func TestInterpolate(t *testing.T) {
	tests := []struct {
		name string
		in   string
		env  *env.Env
		want string
	}{
		{
			name: "replaces_placeholders_from_all_layers",
			in:   "a={{A}} b={{B}} c={{C}}",
			env: &env.Env{
				Shell:   map[string]string{"A": "shellA"},
				Session: map[string]string{"B": "sessionB"},
				File:    map[string]string{"C": "fileC"},
			},
			want: "a=shellA b=sessionB c=fileC",
		},
		{
			name: "leaves_unknown_placeholders_untouched",
			in:   "known={{A}} unknown={{MISSING}}",
			env: &env.Env{
				Shell: map[string]string{"A": "x"},
			},
			want: "known=x unknown={{MISSING}}",
		},
		{
			name: "replaces_multiple_occurrences",
			in:   "{{A}}-{{A}}-{{A}}",
			env: &env.Env{
				Shell: map[string]string{"A": "x"},
			},
			want: "x-x-x",
		},
		{
			// Deterministic: each map contains exactly one key/value and
			// Shell replacement runs before Session replacement.
			name: "nested_interpolation_across_layers",
			in:   "value={{A}}",
			env: &env.Env{
				Shell:   map[string]string{"A": "{{B}}"},
				Session: map[string]string{"B": "final"},
			},
			want: "value=final",
		},
		{
			name: "handles_nil_maps",
			in:   "value={{A}}",
			env:  &env.Env{},
			want: "value={{A}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := interpolate(tt.in, tt.env)
			if got != tt.want {
				t.Fatalf("interpolate(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestExecute(t *testing.T) {
	tests := []struct {
		name      string
		ctx       func(t *testing.T) context.Context
		req       func(t *testing.T) (collection.Request, *env.Env)
		wantErr   bool
		errSubstr string
		check     func(t *testing.T, res Result)
	}{
		{
			name: "empty_url_fails",
			req: func(t *testing.T) (collection.Request, *env.Env) {
				t.Helper()
				return collection.Request{Method: "GET", URL: "   "}, env.NewFromShell()
			},
			wantErr:   true,
			errSubstr: "url is required",
		},
		{
			name: "unresolved_placeholder_in_url_fails",
			req: func(t *testing.T) (collection.Request, *env.Env) {
				t.Helper()
				return collection.Request{Method: "GET", URL: "https://example.com/{{MISSING}}"}, env.NewFromShell()
			},
			wantErr:   true,
			errSubstr: "unresolved variable",
		},
		{
			name: "get_success_returns_status_body_and_trace",
			req: func(t *testing.T) (collection.Request, *env.Env) {
				t.Helper()
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodGet {
						http.Error(w, "want GET", http.StatusMethodNotAllowed)
						return
					}
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("ok"))
				}))
				t.Cleanup(srv.Close)
				return collection.Request{Method: http.MethodGet, URL: srv.URL + "/"}, env.NewFromShell()
			},
			check: func(t *testing.T, res Result) {
				t.Helper()
				if res.Status != http.StatusOK {
					t.Fatalf("Status = %d, want %d", res.Status, http.StatusOK)
				}
				if string(res.Body) != "ok" {
					t.Fatalf("Body = %q, want %q", res.Body, "ok")
				}
				if res.Size != 2 {
					t.Fatalf("Size = %d, want 2", res.Size)
				}
				if len(res.Trace.Events) == 0 {
					t.Fatal("expected trace events")
				}
				if res.Trace.Events[0].Name != "request started" {
					t.Fatalf("first event = %q, want request started", res.Trace.Events[0].Name)
				}
				if res.Trace.RemoteAddr == "" {
					t.Fatal("expected RemoteAddr from GotConn")
				}
				if res.Proto == "" {
					t.Fatal("expected HTTP response protocol")
				}
			},
		},
		{
			name: "post_success_sends_body",
			req: func(t *testing.T) (collection.Request, *env.Env) {
				t.Helper()
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodPost {
						http.Error(w, "want POST", http.StatusMethodNotAllowed)
						return
					}
					b, err := io.ReadAll(r.Body)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					if string(b) != `{"a":1}` {
						http.Error(w, "bad body", http.StatusBadRequest)
						return
					}
					w.WriteHeader(http.StatusCreated)
					_, _ = w.Write([]byte("created"))
				}))
				t.Cleanup(srv.Close)
				return collection.Request{
					Method: http.MethodPost,
					URL:    srv.URL + "/create",
					Body:   `{"a":1}`,
				}, env.NewFromShell()
			},
			check: func(t *testing.T, res Result) {
				t.Helper()
				if res.Status != http.StatusCreated {
					t.Fatalf("Status = %d, want %d", res.Status, http.StatusCreated)
				}
				if string(res.Body) != "created" {
					t.Fatalf("Body = %q", res.Body)
				}
			},
		},
		{
			name: "response_over_limit_sets_response_too_big",
			req: func(t *testing.T) (collection.Request, *env.Env) {
				t.Helper()
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					chunk := make([]byte, 1024)
					for i := range chunk {
						chunk[i] = 'a'
					}
					for written := 0; written < maxResponseBodySize+1; written += len(chunk) {
						_, _ = w.Write(chunk)
					}
				}))
				t.Cleanup(srv.Close)
				return collection.Request{Method: http.MethodGet, URL: srv.URL + "/"}, env.NewFromShell()
			},
			check: func(t *testing.T, res Result) {
				t.Helper()
				if !res.ResponseTooBig {
					t.Fatal("ResponseTooBig = false, want true")
				}
				if res.Body != nil {
					t.Fatalf("Body should be nil when ResponseTooBig, got %d bytes", len(res.Body))
				}
				if res.ResponseSize <= 0 {
					t.Fatalf("ResponseSize = %d, want > 0", res.ResponseSize)
				}
			},
		},
		{
			name: "context_cancellation_returns_error",
			ctx: func(t *testing.T) context.Context {
				t.Helper()
				ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
				t.Cleanup(cancel)
				return ctx
			},
			req: func(t *testing.T) (collection.Request, *env.Env) {
				t.Helper()
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// stall indefinitely — never writes a response
					<-r.Context().Done()
				}))
				t.Cleanup(srv.Close)
				return collection.Request{Method: http.MethodGet, URL: srv.URL + "/"}, env.NewFromShell()
			},
			wantErr: true,
		},
		{
			name: "redirect_appends_next_url_to_trace_redirects",
			req: func(t *testing.T) (collection.Request, *env.Env) {
				t.Helper()
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/":
						http.Redirect(w, r, "/final", http.StatusFound)
					case "/final":
						w.WriteHeader(http.StatusOK)
						_, _ = w.Write([]byte("final"))
					default:
						http.NotFound(w, r)
					}
				}))
				t.Cleanup(srv.Close)
				return collection.Request{Method: http.MethodGet, URL: srv.URL + "/"}, env.NewFromShell()
			},
			check: func(t *testing.T, res Result) {
				t.Helper()
				if res.Status != http.StatusOK {
					t.Fatalf("Status = %d", res.Status)
				}
				if string(res.Body) != "final" {
					t.Fatalf("Body = %q", res.Body)
				}
				if len(res.Trace.Redirects) != 1 {
					t.Fatalf("Redirects = %#v, want 1 entry", res.Trace.Redirects)
				}
				if !strings.HasSuffix(res.Trace.Redirects[0], "/final") {
					t.Fatalf("Redirects[0] = %q, want suffix /final", res.Trace.Redirects[0])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.ctx != nil {
				ctx = tt.ctx(t)
			}
			req, e := tt.req(t)
			res, err := Execute(ctx, req, e)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Fatalf("error = %q, want substring %q", err.Error(), tt.errSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Execute: %v", err)
			}
			if tt.check != nil {
				tt.check(t, res)
			}
		})
	}
}
