package store

import "testing"

func openTestStore(t *testing.T) *Store {
	t.Helper()
	st, err := Open(t.TempDir() + "/t.db")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

func TestAddAndReadTags(t *testing.T) {
	st := openTestStore(t)
	if err := st.AddTag("s1", "Kish", "manual"); err != nil { // mixed case → normalized
		t.Fatal(err)
	}
	if err := st.AddTag("s1", "cost-guard", "auto"); err != nil {
		t.Fatal(err)
	}
	tags, err := st.TagsFor("s1")
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 2 || tags[0].Tag != "cost-guard" || tags[1].Tag != "kish" {
		t.Fatalf("unexpected tags: %+v", tags)
	}
}

func TestAddTagIdempotent(t *testing.T) {
	st := openTestStore(t)
	_ = st.AddTag("s1", "godot", "auto")
	_ = st.AddTag("s1", "godot", "auto")
	if c, _ := st.TagCounts(); c["godot"] != 1 {
		t.Fatalf("godot should count once, got %d", c["godot"])
	}
}

func TestTagCountsAndSessionsByTag(t *testing.T) {
	st := openTestStore(t)
	_ = st.AddTag("s1", "kish", "auto")
	_ = st.AddTag("s2", "kish", "auto")
	_ = st.AddTag("s2", "godot", "auto")
	counts, _ := st.TagCounts()
	if counts["kish"] != 2 || counts["godot"] != 1 {
		t.Fatalf("counts wrong: %+v", counts)
	}
	ids, _ := st.SessionsByTag("kish")
	if len(ids) != 2 {
		t.Fatalf("kish should have 2 sessions, got %v", ids)
	}
}

// The core failure-mode guarantee: re-running a source must not touch tags of
// another source (manual tags survive a heuristic/Lisa re-run).
func TestReplaceTagsBySourcePreservesOtherSources(t *testing.T) {
	st := openTestStore(t)
	_ = st.AddTag("s1", "important", "manual")
	_ = st.AddTag("s1", "old-auto", "auto")

	// Re-run the heuristic: replace only the "auto" tags.
	if err := st.ReplaceTagsBySource("s1", "auto", []string{"new-auto", "kish"}); err != nil {
		t.Fatal(err)
	}
	tags, _ := st.TagsFor("s1")
	got := map[string]string{}
	for _, tg := range tags {
		got[tg.Tag] = tg.Source
	}
	if got["important"] != "manual" {
		t.Error("manual tag must survive an auto re-run")
	}
	if _, ok := got["old-auto"]; ok {
		t.Error("old auto tag should be gone")
	}
	if got["new-auto"] != "auto" || got["kish"] != "auto" {
		t.Errorf("new auto tags missing: %+v", got)
	}
}

func TestReplaceTagsDedupesAndNormalizes(t *testing.T) {
	st := openTestStore(t)
	_ = st.ReplaceTagsBySource("s1", "auto", []string{"Kish", "kish", " KISH ", ""})
	if c, _ := st.TagCounts(); c["kish"] != 1 {
		t.Fatalf("duplicates/case/blank should collapse to one, got %d", c["kish"])
	}
}

func TestRemoveTag(t *testing.T) {
	st := openTestStore(t)
	_ = st.AddTag("s1", "tmp", "manual")
	if err := st.RemoveTag("s1", "TMP"); err != nil { // case-insensitive
		t.Fatal(err)
	}
	if tags, _ := st.TagsFor("s1"); len(tags) != 0 {
		t.Fatalf("tag should be removed, got %+v", tags)
	}
}
