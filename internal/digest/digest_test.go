package digest_test

import (
	"testing"

	"portwatch/internal/digest"
)

func entries() []digest.Entry {
	return []digest.Entry{
		{Protocol: "tcp", Addr: "0.0.0.0", Port: 8080, Inode: 1001},
		{Protocol: "tcp", Addr: "127.0.0.1", Port: 443, Inode: 1002},
	}
}

func TestUpdate_FirstCallAlwaysChanged(t *testing.T) {
	t.Parallel()
	tr := digest.New()
	changed, d := tr.Update(entries())
	if !changed {
		t.Fatal("expected changed=true on first call")
	}
	if d == "" {
		t.Fatal("expected non-empty digest")
	}
}

func TestUpdate_SameEntriesNoChange(t *testing.T) {
	t.Parallel()
	tr := digest.New()
	tr.Update(entries())
	changed, _ := tr.Update(entries())
	if changed {
		t.Fatal("expected changed=false for identical entries")
	}
}

func TestUpdate_OrderIndependent(t *testing.T) {
	t.Parallel()
	tr := digest.New()
	e := entries()
	_, d1 := tr.Update(e)
	// reverse order
	reversed := []digest.Entry{e[1], e[0]}
	_, d2 := tr.Update(reversed)
	if d1 != d2 {
		t.Fatalf("digest should be order-independent: %s != %s", d1, d2)
	}
}

func TestUpdate_DifferentEntriesChanged(t *testing.T) {
	t.Parallel()
	tr := digest.New()
	tr.Update(entries())
	newEntries := []digest.Entry{
		{Protocol: "udp", Addr: "0.0.0.0", Port: 53, Inode: 2001},
	}
	changed, _ := tr.Update(newEntries)
	if !changed {
		t.Fatal("expected changed=true when entries differ")
	}
}

func TestUpdate_EmptyEntries(t *testing.T) {
	t.Parallel()
	tr := digest.New()
	changed, d := tr.Update(nil)
	if !changed {
		t.Fatal("expected changed=true on first call with empty entries")
	}
	changed2, d2 := tr.Update([]digest.Entry{})
	if changed2 {
		t.Fatal("expected changed=false for repeated empty snapshot")
	}
	if d != d2 {
		t.Fatalf("empty digests should match: %s != %s", d, d2)
	}
}

func TestLast_ReturnsLatestDigest(t *testing.T) {
	t.Parallel()
	tr := digest.New()
	if tr.Last() != "" {
		t.Fatal("expected empty Last() before any Update")
	}
	_, d := tr.Update(entries())
	if tr.Last() != d {
		t.Fatalf("Last() = %s; want %s", tr.Last(), d)
	}
}
