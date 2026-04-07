package auth

import (
	"testing"

	"github.com/jxdones/ferret/internal/collection"
)

func TestResolve(t *testing.T) {
	def := &collection.Auth{Type: "bearer", Token: "default"}
	reqBearer := &collection.Auth{Type: "bearer", Token: "req"}
	reqNone := &collection.Auth{Type: "none"}
	reqInherit := &collection.Auth{Type: "inherit"}
	reqEmptyType := &collection.Auth{Type: ""}

	tests := []struct {
		name string
		req  *collection.Auth
		def  *collection.Auth
		want *collection.Auth
	}{
		{
			name: "when_req_nil_returns_def",
			req:  nil,
			def:  def,
			want: def,
		},
		{
			name: "when_req_nil_and_def_nil_returns_nil",
			req:  nil,
			def:  nil,
			want: nil,
		},
		{
			name: "when_req_type_empty_returns_def",
			req:  reqEmptyType,
			def:  def,
			want: def,
		},
		{
			name: "when_req_type_inherit_returns_def",
			req:  reqInherit,
			def:  def,
			want: def,
		},
		{
			name: "when_req_type_none_returns_nil",
			req:  reqNone,
			def:  def,
			want: nil,
		},
		{
			name: "when_req_type_explicit_returns_req",
			req:  reqBearer,
			def:  def,
			want: reqBearer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Resolve(tt.req, tt.def)
			if got != tt.want {
				t.Fatalf("Resolve(%#v, %#v) = %#v, want %#v", tt.req, tt.def, got, tt.want)
			}
		})
	}
}
