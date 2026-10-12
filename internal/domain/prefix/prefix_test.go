package prefix

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	valid := []struct {
		name string
		in   string
		want Prefix
	}{
		{"trailing slash", "/raspberry/", "/raspberry/"},
		{"no trailing slash", "/raspberry", "/raspberry/"},
		{"single letter", "/a/", "/a/"},
		{"all digits", "/100/", "/100/"},
		{"interior hyphens", "/home-lab-2/", "/home-lab-2/"},
		{"max length", "/" + strings.Repeat("a", maxNodeLen) + "/", Prefix("/" + strings.Repeat("a", maxNodeLen) + "/")},
	}
	for _, tt := range valid {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := Parse(tt.in)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v; want nil", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("Parse(%q) = %q; want %q", tt.in, got, tt.want)
			}
		})
	}

	invalid := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"root", "/"},
		{"no leading slash", "raspberry/"},
		{"bare name", "raspberry"},
		{"double leading slash", "//raspberry/"},
		{"multiple segments", "/raspberry/docs/"},
		{"double trailing slash", "/raspberry//"},
		{"uppercase", "/Raspberry/"},
		{"reserved control path", "/_relay/"},
		{"underscore", "/my_node/"},
		{"leading hyphen", "/-raspberry/"},
		{"trailing hyphen", "/raspberry-/"},
		{"dot", "/rasp.berry/"},
		{"space", "/rasp berry/"},
		{"too long", "/" + strings.Repeat("a", maxNodeLen+1) + "/"},
	}
	for _, tt := range invalid {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if _, err := Parse(tt.in); !errors.Is(err, ErrInvalid) {
				t.Errorf("Parse(%q) error = %v; want %v", tt.in, err, ErrInvalid)
			}
		})
	}
}

func TestStringAndNode(t *testing.T) {
	p, err := Parse("/raspberry")
	if err != nil {
		t.Fatalf("Parse(): %v", err)
	}
	if got := p.String(); got != "/raspberry/" {
		t.Errorf("String() = %q; want %q", got, "/raspberry/")
	}
	if got := p.Node(); got != "raspberry" {
		t.Errorf("Node() = %q; want %q", got, "raspberry")
	}
}

func TestContains(t *testing.T) {
	p, err := Parse("/raspberry/")
	if err != nil {
		t.Fatalf("Parse(): %v", err)
	}
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"under", "/raspberry/docs/", true},
		{"exact", "/raspberry/", true},
		{"bare node root", "/raspberry", true},
		{"sibling sharing name", "/raspberry-evil/", false},
		{"other node", "/laptop/", false},
		{"root", "/", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := p.Contains(tt.path); got != tt.want {
				t.Errorf("Contains(%q) = %v; want %v", tt.path, got, tt.want)
			}
		})
	}
}

// ExampleParse shows the canonical form of a node prefix.
func ExampleParse() {
	p, err := Parse("/raspberry")
	if err != nil {
		panic(err)
	}
	fmt.Println(p.String(), p.Node())
	// Output:
	// /raspberry/ raspberry
}
