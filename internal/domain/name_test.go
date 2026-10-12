package domain_test

import (
	"strings"
	"testing"

	"relay/internal/domain"
)

// AC-10: name grammar (contract section 2.2):
// [a-z0-9]([a-z0-9-]{0,61}[a-z0-9])? with _relay reserved.
func TestParseName(t *testing.T) {
	valid := "a"
	valid63 := "a" + strings.Repeat("b", 61) + "c" // 63 chars, the max

	cases := []struct {
		in      string
		wantErr bool
	}{
		{in: valid, wantErr: false},
		{in: "mynode", wantErr: false},
		{in: "abc-123", wantErr: false},
		{in: "a-b-c", wantErr: false},
		{in: "0", wantErr: false},
		{in: valid63, wantErr: false},
		{in: "", wantErr: true},
		{in: "MyNode", wantErr: true}, // uppercase
		{in: "-abc", wantErr: true},   // leading hyphen
		{in: "abc-", wantErr: true},   // trailing hyphen
		{in: "a_b", wantErr: true},    // underscore
		{in: "a.b", wantErr: true},    // dot
		{in: "a b", wantErr: true},    // space
		{in: valid63 + "d", wantErr: true},
		{in: "_relay", wantErr: true}, // reserved node name
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			_, err := domain.ParseName(tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ParseName(%q) error = %v, want error = %v", tc.in, err, tc.wantErr)
			}
		})
	}
}
