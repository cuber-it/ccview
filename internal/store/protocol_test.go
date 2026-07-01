package store

import "testing"

func TestProtocolDefaultsToZero(t *testing.T) {
	st := openTestStore(t)
	ps, err := st.GetProtocol("nope")
	if err != nil {
		t.Fatal(err)
	}
	if ps.Active || ps.EventsWritten != 0 || ps.PromptNum != 0 {
		t.Fatalf("unrecorded session should be zero, got %+v", ps)
	}
}

func TestProtocolActiveAndMarkerAreIndependent(t *testing.T) {
	st := openTestStore(t)
	// Marker set while off, then turned on: neither write clobbers the other.
	if err := st.SetProtocolMarker("s1", 120, 7); err != nil {
		t.Fatal(err)
	}
	if err := st.SetProtocolActive("s1", true); err != nil {
		t.Fatal(err)
	}
	ps, _ := st.GetProtocol("s1")
	if !ps.Active || ps.EventsWritten != 120 || ps.PromptNum != 7 {
		t.Fatalf("active flip must preserve marker, got %+v", ps)
	}

	// Turning off must keep the resume marker so on-again continues.
	if err := st.SetProtocolActive("s1", false); err != nil {
		t.Fatal(err)
	}
	ps, _ = st.GetProtocol("s1")
	if ps.Active {
		t.Fatal("should be inactive")
	}
	if ps.EventsWritten != 120 || ps.PromptNum != 7 {
		t.Fatalf("marker must survive off, got %+v", ps)
	}
}

func TestActiveProtocols(t *testing.T) {
	st := openTestStore(t)
	_ = st.SetProtocolActive("s1", true)
	_ = st.SetProtocolActive("s2", false)
	_ = st.SetProtocolActive("s3", true)
	ids, err := st.ActiveProtocols()
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 active, got %v", ids)
	}
}

func TestDeleteMetaRemovesProtocol(t *testing.T) {
	st := openTestStore(t)
	_ = st.SetProtocolActive("s1", true)
	if err := st.DeleteMeta("s1"); err != nil {
		t.Fatal(err)
	}
	if ps, _ := st.GetProtocol("s1"); ps.Active {
		t.Fatal("protocol row should be gone after DeleteMeta")
	}
}
