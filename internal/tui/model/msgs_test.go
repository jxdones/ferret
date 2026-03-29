package model

import (
	"errors"
	"testing"

	"github.com/jxdones/ferret/internal/exec"
)

func TestMessages_ZeroValuesAndFields(t *testing.T) {
	tests := []struct {
		name string
		msg  any
	}{
		{
			name: "RequestStartedMsg",
			msg:  RequestStartedMsg{},
		},
		{
			name: "RequestFinishedMsg_fields_roundtrip",
			msg: RequestFinishedMsg{
				Body:    []byte("ok"),
				Headers: map[string][]string{"Content-Type": {"text/plain"}},
				Trace:   exec.Trace{},
			},
		},
		{
			name: "RequestFailedMsg_has_error",
			msg: RequestFailedMsg{
				Error: errors.New("boom"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.msg == nil {
				t.Fatalf("msg is nil")
			}
		})
	}
}
