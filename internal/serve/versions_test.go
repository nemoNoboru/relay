package serve_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nemoNoboru/relay/internal/serve"
)

// writeLive writes a site's live directory under <data>/sites/<name>.
func writeLive(t *testing.T, data, name, content string) {
	t.Helper()
	dir := filepath.Join(data, "sites", name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readLive(t *testing.T, data, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(data, "sites", name, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// TestVersionsSnapshotList pins that Snapshot moves the live site into a
// version and List reports it.
func TestVersionsSnapshotList(t *testing.T) {
	data := t.TempDir()
	v := serve.NewVersions(data)

	writeLive(t, data, "blog", "v1")
	id, err := v.Snapshot("blog")
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Fatalf("snapshot id is empty")
	}
	if _, err := os.Stat(filepath.Join(data, "sites", "blog")); !os.IsNotExist(err) {
		t.Fatalf("live site still present after snapshot")
	}
	vs, err := v.List("blog")
	if err != nil {
		t.Fatal(err)
	}
	if len(vs) != 1 || vs[0].ID != id {
		t.Fatalf("versions = %+v, want one entry %q", vs, id)
	}
}

// TestVersionsRollback pins that Rollback restores an earlier snapshot and
// preserves the displaced live site as a new version.
func TestVersionsRollback(t *testing.T) {
	data := t.TempDir()
	v := serve.NewVersions(data)

	writeLive(t, data, "blog", "v1")
	id1, err := v.Snapshot("blog")
	if err != nil {
		t.Fatal(err)
	}
	writeLive(t, data, "blog", "v2")
	if _, err := v.Snapshot("blog"); err != nil {
		t.Fatal(err)
	}
	writeLive(t, data, "blog", "v3") // current live

	if _, err := v.Rollback("blog", id1); err != nil {
		t.Fatal(err)
	}
	if got := readLive(t, data, "blog"); got != "v1" {
		t.Fatalf("live content after rollback = %q, want v1", got)
	}
	// The rolled-back-to version left the list; the displaced v3 joined it.
	vs, err := v.List("blog")
	if err != nil {
		t.Fatal(err)
	}
	if len(vs) != 2 {
		t.Fatalf("versions after rollback = %d, want 2 (v2 + displaced v3)", len(vs))
	}

	if _, err := v.Rollback("blog", "dk_nope"); err == nil {
		t.Fatalf("rollback to an unknown version succeeded")
	}
}

// TestVersionsPrune pins that Prune trims beyond the retention cap.
func TestVersionsPrune(t *testing.T) {
	data := t.TempDir()
	v := serve.NewVersions(data)

	for i := 0; i < 12; i++ {
		writeLive(t, data, "blog", "x")
		if _, err := v.Snapshot("blog"); err != nil {
			t.Fatal(err)
		}
	}
	if err := v.Prune("blog"); err != nil {
		t.Fatal(err)
	}
	vs, err := v.List("blog")
	if err != nil {
		t.Fatal(err)
	}
	if len(vs) != 10 {
		t.Fatalf("versions after prune = %d, want 10", len(vs))
	}
}
