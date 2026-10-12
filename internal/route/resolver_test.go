package route_test

import (
	"errors"
	"testing"

	"relay/internal/domain"
	"relay/internal/route"
)

// AC-11: route resolution (contract section 5).
// <node> = first segment, <site> = second, Rest = the remainder, rooted.
func TestResolve(t *testing.T) {
	cases := []struct {
		path   string
		want   *route.Target
		errErr error // exact expected sentinel error
	}{
		{path: "/mynode/docs/", want: &route.Target{Node: "mynode", Site: "docs", Rest: "/"}},
		{path: "/mynode/docs", want: &route.Target{Node: "mynode", Site: "docs", Rest: "/"}},
		{path: "/mynode/docs/assets/x.css", want: &route.Target{Node: "mynode", Site: "docs", Rest: "/assets/x.css"}},
		{path: "/mynode/docs/sub/page", want: &route.Target{Node: "mynode", Site: "docs", Rest: "/sub/page"}},
		{path: "/mynode/docs/sub/", want: &route.Target{Node: "mynode", Site: "docs", Rest: "/sub/"}},
		{path: "/a//b", want: &route.Target{Node: "a", Site: "b", Rest: "/"}}, // // collapses (CleanPath)
		{path: "//mynode/docs/", want: &route.Target{Node: "mynode", Site: "docs", Rest: "/"}},
		{path: "/mynode", errErr: route.ErrNoSite},
		{path: "/", errErr: route.ErrNoNode},
		{path: "/_relay/v1/node", errErr: route.ErrReserved},
		{path: "/Mynode/docs/", errErr: route.ErrNoNode}, // grammar-invalid => unknown node, not an error of its own
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			got, err := route.Resolve(tc.path)
			if tc.errErr != nil {
				if !errors.Is(err, tc.errErr) {
					t.Fatalf("Resolve(%q) error = %v, want %v", tc.path, err, tc.errErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Resolve(%q) unexpected error: %v", tc.path, err)
			}
			if got == nil {
				t.Fatalf("Resolve(%q) returned nil target", tc.path)
			}
			if got.Node != tc.want.Node || got.Site != tc.want.Site || got.Rest != tc.want.Rest {
				t.Fatalf("Resolve(%q) = {%s %s %s}, want {%s %s %s}",
					tc.path, got.Node, got.Site, got.Rest, tc.want.Node, tc.want.Site, tc.want.Rest)
			}
		})
	}
}

// The reserved-name sentinel must be the same value domain.ParseName uses,
// so handler and resolver cannot drift apart.
func TestReservedIsDomainSentinel(t *testing.T) {
	if !errors.Is(route.ErrReserved, domain.ErrReservedName) {
		t.Fatalf("route.ErrReserved = %v, want it to be/is domain.ErrReservedName (%v)", route.ErrReserved, domain.ErrReservedName)
	}
}
