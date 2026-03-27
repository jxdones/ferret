package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/jxdones/ferret/internal/collection"
	"github.com/jxdones/ferret/internal/exec"
)

// RawBody writes only the HTTP response body to w (no summary line, no re-indenting).
// Use this when piping to jq or other JSON tools.
func RawBody(res exec.Result, w io.Writer) error {
	_, err := w.Write(res.Body)
	return err
}

// Response writes a formatted response to w, including a summary header line and the response body.
func Response(req collection.Request, res exec.Result, w io.Writer) error {
	_, err := fmt.Fprintf(w, "%s  ·  %s %s  ·  %s  ·  %dms  ·  %db\n",
		req.Name,
		req.Method,
		res.URL,
		res.StatusText,
		res.Duration.Milliseconds(),
		res.Size,
	)
	if err != nil {
		return err
	}
	jsonPrettyPrint(res, w)
	return nil
}

// jsonPrettyPrint pretty prints the response body as JSON.
func jsonPrettyPrint(res exec.Result, w io.Writer) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, res.Body, "", "  "); err != nil {
		w.Write(res.Body)
		return
	}
	w.Write(buf.Bytes())
}
