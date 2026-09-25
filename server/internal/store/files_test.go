package store

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"modders-feedback/server/internal/domain"
)

func TestFilesRoundTripAndRecompress(t *testing.T) {
	root := t.TempDir()
	files, err := NewFiles(filepath.Join(root, "attachments"))
	if err != nil {
		t.Fatal(err)
	}
	raw := bytes.Repeat([]byte("screenshot-bytes-"), 4000)
	const id = "0123456789abcdef0123456789abcdef"
	if err = files.Save(id, raw, domain.StatusOpen); err != nil {
		t.Fatal(err)
	}
	fast, err := os.ReadFile(files.objectPath(id))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(fast, []byte("fast\n")) {
		t.Fatalf("active archive label = %q", fast[:8])
	}
	opened, err := files.Open(id)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(opened, raw) {
		t.Fatal("opened bytes differ")
	}
	if err = files.Save(id, raw, domain.StatusResolved); err != nil {
		t.Fatal(err)
	}
	maxed, err := os.ReadFile(files.objectPath(id))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(maxed, []byte("max\n")) {
		t.Fatal("resolved archive is not max")
	}
	if len(maxed) > len(fast) {
		t.Fatalf("resolved archive grew: %d > %d", len(maxed), len(fast))
	}
	opened, err = files.Open(id)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(opened, raw) {
		t.Fatal("recompressed bytes differ")
	}
	if err = files.Delete(id); err != nil {
		t.Fatal(err)
	}
	if _, err = files.Open(id); !os.IsNotExist(err) {
		t.Fatalf("deleted file open err = %v", err)
	}
}

func TestFilesRejectsRelativeRoot(t *testing.T) {
	if _, err := NewFiles("relative/path"); err == nil {
		t.Fatal("relative root accepted")
	}
}
