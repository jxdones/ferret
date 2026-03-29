package render

import (
	"bytes"
	"testing"

	"github.com/jxdones/ferret/internal/exec"
)

func TestRawBodyWritesRawResponse(t *testing.T) {
	res := exec.Result{
		Proto:      "HTTP/1.1",
		StatusText: "200 OK",
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
			"Date":         {"Sun, 29 Mar 2026 05:19:27 GMT"},
		},
		Body: []byte(`{"success":"true"}`),
	}
	var out bytes.Buffer
	if err := RawBody(res, &out); err != nil {
		t.Fatalf("RawBody: %v", err)
	}
	want := "HTTP/1.1 200 OK\nContent-Type: application/json\nDate: Sun, 29 Mar 2026 05:19:27 GMT\n\n{\"success\":\"true\"}"
	if out.String() != want {
		t.Fatalf("output = %q, want %q", out.String(), want)
	}
}

func TestResponsePrettyPrintsJSON(t *testing.T) {
	res := exec.Result{Body: []byte(`{"ok":true}`)}
	var out bytes.Buffer
	if err := Response(res, &out); err != nil {
		t.Fatalf("Response: %v", err)
	}
	want := "{\n  \"ok\": true\n}"
	if out.String() != want {
		t.Fatalf("output = %q, want %q", out.String(), want)
	}
}

func TestResponseWritesPlainTextBody(t *testing.T) {
	res := exec.Result{Body: []byte("plain text")}
	var out bytes.Buffer
	if err := Response(res, &out); err != nil {
		t.Fatalf("Response: %v", err)
	}
	if out.String() != "plain text" {
		t.Fatalf("output = %q, want %q", out.String(), "plain text")
	}
}
