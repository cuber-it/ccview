package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreRoundtrip(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	// notes: empty default, set, upsert
	if n, _ := st.GetNote("s1"); n != "" {
		t.Errorf("unknown note want %q, got %q", "", n)
	}
	if err := st.SetNote("s1", "hello"); err != nil {
		t.Fatal(err)
	}
	if err := st.SetNote("s1", "updated"); err != nil {
		t.Fatal(err)
	}
	if n, _ := st.GetNote("s1"); n != "updated" {
		t.Errorf("note upsert want %q, got %q", "updated", n)
	}

	// roots: unset → set → read
	if _, ok, _ := st.GetRoots(); ok {
		t.Error("roots should be unset initially")
	}
	if err := st.SetRoots([]string{"/a", "/b"}); err != nil {
		t.Fatal(err)
	}
	if r, ok, _ := st.GetRoots(); !ok || len(r) != 2 || r[0] != "/a" {
		t.Errorf("roots want [/a /b], got %v", r)
	}

	// name + favorite
	if err := st.SetName("s1", "Mein Name"); err != nil {
		t.Fatal(err)
	}
	if err := st.SetFavorite("s1", true); err != nil {
		t.Fatal(err)
	}
	if err := st.SetName("s2", "Andere"); err != nil {
		t.Fatal(err)
	}
	meta, _ := st.AllMeta()
	if meta["s1"].Name != "Mein Name" || !meta["s1"].Favorite {
		t.Errorf("s1 meta wrong: %+v", meta["s1"])
	}
	if meta["s2"].Name != "Andere" || meta["s2"].Favorite {
		t.Errorf("s2 meta wrong: %+v", meta["s2"])
	}

	// groups: save, upsert
	if err := st.SaveGroups([]Group{
		{Key: "p1", Label: "Projekt 1", Order: 0},
		{Key: "p2", Label: "P2", Order: 1, Hidden: true},
	}); err != nil {
		t.Fatal(err)
	}
	if err := st.SaveGroups([]Group{{Key: "p1", Label: "Renamed", Order: 5, Hidden: true}}); err != nil {
		t.Fatal(err)
	}
	gm := map[string]Group{}
	gs, _ := st.Groups()
	for _, g := range gs {
		gm[g.Key] = g
	}
	if gm["p1"].Label != "Renamed" || gm["p1"].Order != 5 || !gm["p1"].Hidden {
		t.Errorf("p1 upsert wrong: %+v", gm["p1"])
	}
}

func TestMigrateFromFiles(t *testing.T) {
	dir := t.TempDir()
	rootsPath := filepath.Join(dir, "roots.json")
	os.WriteFile(rootsPath, []byte(`{"roots":["/x","/y"]}`), 0o644)
	notesDir := filepath.Join(dir, "notes")
	os.MkdirAll(notesDir, 0o755)
	os.WriteFile(filepath.Join(notesDir, "abc.md"), []byte("note abc"), 0o644)
	os.WriteFile(filepath.Join(notesDir, "empty.md"), []byte(""), 0o644)

	st, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	if err := st.MigrateFromFiles(rootsPath, notesDir); err != nil {
		t.Fatal(err)
	}
	if r, ok, _ := st.GetRoots(); !ok || len(r) != 2 || r[0] != "/x" {
		t.Errorf("migrated roots wrong: %v", r)
	}
	if n, _ := st.GetNote("abc"); n != "note abc" {
		t.Errorf("migrated note wrong: %q", n)
	}
	if n, _ := st.GetNote("empty"); n != "" {
		t.Errorf("empty note should stay empty, got %q", n)
	}

	// idempotent: must NOT overwrite existing state
	st.SetRoots([]string{"/keep"})
	st.SetNote("abc", "manually changed")
	os.WriteFile(rootsPath, []byte(`{"roots":["/should","/not","/win"]}`), 0o644)
	if err := st.MigrateFromFiles(rootsPath, notesDir); err != nil {
		t.Fatal(err)
	}
	if r, _, _ := st.GetRoots(); len(r) != 1 || r[0] != "/keep" {
		t.Errorf("migration overwrote existing roots: %v", r)
	}
	if n, _ := st.GetNote("abc"); n != "manually changed" {
		t.Errorf("migration overwrote existing note: %q", n)
	}
}
