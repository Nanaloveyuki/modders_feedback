package store

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"syscall"
	"testing"
	"time"
)

func TestBackupKeepsTenNewestCopies(t *testing.T) {
	dir := t.TempDir()
	db := openPopulated(t, dir)
	backupDir := filepath.Join(dir, backupDirName)
	if err := os.MkdirAll(backupDir, backupDirPerm); err != nil {
		t.Fatal(err)
	}
	oldest := filepath.Join(backupDir, "feedback-20200101T000000Z.db")
	for hour := 0; hour < backupKeep; hour++ {
		name := time.Date(2020, 1, 1, hour, 0, 0, 0, time.UTC).Format("20060102T150405Z")
		if err := os.WriteFile(filepath.Join(backupDir, "feedback-"+name+".db"), []byte("old"), backupFilePerm); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Backup(context.Background()); err != nil {
		t.Fatal(err)
	}
	names := backupNames(t, backupDir)
	if len(names) != backupKeep {
		t.Fatalf("kept %d backups, want %d: %v", len(names), backupKeep, names)
	}
	if _, err := os.Stat(oldest); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("oldest backup still present: %v", err)
	}
	newest := names[len(names)-1]
	restored := openRestored(t, filepath.Join(backupDir, newest))
	var count int
	if err := restored.db.QueryRow(`SELECT COUNT(*) FROM notes`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("restored rows = %d, want 1", count)
	}
}

func TestBackupDropsOldestCopyWhenDiskIsFull(t *testing.T) {
	dir := t.TempDir()
	db := openPopulated(t, dir)
	backupDir := filepath.Join(dir, backupDirName)
	if err := os.MkdirAll(backupDir, backupDirPerm); err != nil {
		t.Fatal(err)
	}
	oldest := filepath.Join(backupDir, "feedback-20200101T000000Z.db")
	newer := filepath.Join(backupDir, "feedback-20200102T000000Z.db")
	for _, path := range []string{oldest, newer} {
		if err := os.WriteFile(path, []byte("old"), backupFilePerm); err != nil {
			t.Fatal(err)
		}
	}
	original := copyDatabase
	defer func() { copyDatabase = original }()
	calls := 0
	copyDatabase = func(_ context.Context, _ *Store, destination string) error {
		calls++
		if calls == 1 {
			return syscall.ENOSPC
		}
		return os.WriteFile(destination, []byte("sqlite-backup"), backupFilePerm)
	}
	if err := db.Backup(context.Background()); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("copy attempts = %d, want 2", calls)
	}
	if _, err := os.Stat(oldest); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("oldest backup still present: %v", err)
	}
	if _, err := os.Stat(newer); err != nil {
		t.Fatal(err)
	}
	names := backupNames(t, backupDir)
	if len(names) != 2 {
		t.Fatalf("backups = %v, want the newer seed plus one new copy", names)
	}
}

func TestBackupSurvivesPanic(t *testing.T) {
	dir := t.TempDir()
	db := openPopulated(t, dir)
	original := copyDatabase
	defer func() { copyDatabase = original }()
	copyDatabase = func(context.Context, *Store, string) error {
		panic("oom")
	}
	err := db.Backup(context.Background())
	if err == nil || err.Error() != "database backup aborted: oom" {
		t.Fatalf("error = %v", err)
	}
	names := backupNames(t, filepath.Join(dir, backupDirName))
	if len(names) != 0 {
		t.Fatalf("published backups after panic: %v", names)
	}
}

func openPopulated(t *testing.T, dir string) *Store {
	t.Helper()
	db, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err = db.db.Exec(`CREATE TABLE notes(body TEXT); INSERT INTO notes(body) VALUES ('kept')`); err != nil {
		t.Fatal(err)
	}
	return db
}

func openRestored(t *testing.T, path string) *Store {
	t.Helper()
	dir := t.TempDir()
	target := filepath.Join(dir, "feedback.db")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(target, content, 0o600); err != nil {
		t.Fatal(err)
	}
	db, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func backupNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names
}
