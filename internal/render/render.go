package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/jxdones/ferret/internal/exec"
)

// RawBody writes a raw HTTP response (status line, headers, blank line, body).
func RawBody(res exec.Result, w io.Writer) error {
	proto := strings.TrimSpace(res.Proto)
	if proto == "" {
		proto = "HTTP/1.1"
	}
	status := strings.TrimSpace(res.StatusText)
	if status == "" {
		status = fmt.Sprintf("%d", res.Status)
	}
	if _, err := fmt.Fprintf(w, "%s %s\n", proto, status); err != nil {
		return err
	}

	headerKeys := make([]string, 0, len(res.Headers))
	for k := range res.Headers {
		headerKeys = append(headerKeys, k)
	}
	sort.Strings(headerKeys)
	for _, k := range headerKeys {
		for _, v := range res.Headers[k] {
			if _, err := fmt.Fprintf(w, "%s: %s\n", k, v); err != nil {
				return err
			}
		}
	}

	if _, err := io.WriteString(w, "\n"); err != nil {
		return err
	}
	_, err := w.Write(res.Body)
	return err
}

// Response writes only the response body to w. If it's JSON, it is pretty
// printed; otherwise the body is written unchanged.
func Response(res exec.Result, w io.Writer) error {
	return jsonPrettyPrint(res, w)
}

// jsonPrettyPrint pretty prints the response body as JSON.
func jsonPrettyPrint(res exec.Result, w io.Writer) error {
	var buf bytes.Buffer
	if err := json.Indent(&buf, res.Body, "", "  "); err != nil {
		_, writeErr := w.Write(res.Body)
		return writeErr
	}
	_, err := w.Write(buf.Bytes())
	return err
}
