package store

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"sort"
	"strings"
	"syscall"
	"time"

	"modernc.org/sqlite"
)

const (
	backupInterval  = 72 * time.Hour
	backupKeep      = 10
	backupPageBatch = 16
	backupDirName   = "backups"
	backupPrefix    = "feedback-"
	backupSuffix    = ".db"
	backupName      = `^feedback-\d{8}T\d{6}Z\.db$`
	backupFilePerm  = 0o640
	backupDirPerm   = 0o750
)

type sqliteBackup interface {
	Step(n int32) (bool, error)
	Finish() error
}

type sqliteBackuper interface {
	NewBackup(dstURI string) (sqliteBackup, error)
}

type driverBackuper struct {
	raw interface {
		NewBackup(string) (*sqlite.Backup, error)
	}
}

func (d driverBackuper) NewBackup(dstURI string) (sqliteBackup, error) {
	return d.raw.NewBackup(dstURI)
}

func (s *Store) StartBackups(ctx context.Context) {
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("database backup stopped after panic: %v\n%s", recovered, debug.Stack())
			}
		}()
		s.runBackups(ctx)
	}()
}

func (s *Store) runBackups(ctx context.Context) {
	if err := s.Backup(ctx); err != nil {
		log.Printf("database backup failed: %v", err)
	}
	ticker := time.NewTicker(backupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.Backup(ctx); err != nil {
				log.Printf("database backup failed: %v", err)
			}
		}
	}
}

func (s *Store) Backup(ctx context.Context) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("database backup aborted: %v", recovered)
			log.Printf("%v\n%s", err, debug.Stack())
		}
	}()
	if err = ctx.Err(); err != nil {
		return err
	}
	dir := filepath.Join(s.dataDir, backupDirName)
	if err = os.MkdirAll(dir, backupDirPerm); err != nil {
		return err
	}
	finalPath := filepath.Join(dir, backupPrefix+time.Now().UTC().Format("20060102T150405Z")+backupSuffix)
	partialPath := finalPath + ".partial"
	defer func() {
		if err != nil {
			_ = os.Remove(partialPath)
		}
	}()
	if err = copyDatabase(ctx, s, partialPath); err != nil {
		if !isDiskFull(err) {
			return err
		}
		if err = s.retryAfterFreeingSpace(ctx, dir, partialPath, err); err != nil {
			return err
		}
	}
	if err = publishBackup(partialPath, finalPath, dir); err != nil {
		return err
	}
	if _, err = pruneBackups(dir, backupKeep); err != nil {
		return err
	}
	log.Printf("database backup written: %s", finalPath)
	return nil
}

func (s *Store) retryAfterFreeingSpace(ctx context.Context, dir, partialPath string, cause error) error {
	removed, pruneErr := pruneOldestBackup(dir)
	if pruneErr != nil {
		return errors.Join(cause, pruneErr)
	}
	if !removed {
		return cause
	}
	log.Printf("database backup retrying after removing the oldest copy: disk full")
	return copyDatabase(ctx, s, partialPath)
}

func publishBackup(partialPath, finalPath, dir string) error {
	if err := os.Chmod(partialPath, backupFilePerm); err != nil {
		return err
	}
	err := os.Rename(partialPath, finalPath)
	if err == nil || !isDiskFull(err) {
		return err
	}
	removed, pruneErr := pruneOldestBackup(dir)
	if pruneErr != nil {
		return errors.Join(err, pruneErr)
	}
	if !removed {
		return err
	}
	return os.Rename(partialPath, finalPath)
}

var copyDatabase = func(ctx context.Context, s *Store, destination string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	return conn.Raw(func(driverConn any) error {
		raw, ok := driverConn.(interface {
			NewBackup(string) (*sqlite.Backup, error)
		})
		if !ok {
			return errors.New("sqlite connection does not support online backup")
		}
		return copyPages(ctx, driverBackuper{raw}, destination)
	})
}

func copyPages(ctx context.Context, copier sqliteBackuper, destination string) error {
	backup, err := copier.NewBackup(destinationURI(destination))
	if err != nil {
		return err
	}
	for {
		if err = ctx.Err(); err != nil {
			_ = backup.Finish()
			return err
		}
		more, stepErr := backup.Step(backupPageBatch)
		if stepErr != nil {
			_ = backup.Finish()
			return stepErr
		}
		if !more {
			return backup.Finish()
		}
	}
}

func destinationURI(path string) string {
	return "file:" + filepath.ToSlash(path) + "?_pragma=journal_mode(DELETE)"
}

func pruneBackups(dir string, keep int) (bool, error) {
	entries, err := completedBackups(dir)
	if err != nil {
		return false, err
	}
	removed := false
	if keep < 0 {
		keep = 0
	}
	for len(entries) > keep {
		if err = os.Remove(entries[0]); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return removed, err
		}
		removed = true
		entries = entries[1:]
	}
	return removed, nil
}

func pruneOldestBackup(dir string) (bool, error) {
	entries, err := completedBackups(dir)
	if err != nil || len(entries) == 0 {
		return false, err
	}
	err = os.Remove(entries[0])
	if errors.Is(err, fs.ErrNotExist) {
		return true, nil
	}
	return err == nil, err
}

func completedBackups(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	pattern := regexp.MustCompile(backupName)
	completed := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !pattern.MatchString(entry.Name()) {
			continue
		}
		completed = append(completed, filepath.Join(dir, entry.Name()))
	}
	sort.Strings(completed)
	return completed, nil
}

func isDiskFull(err error) bool {
	return errors.Is(err, syscall.ENOSPC) || errors.Is(err, syscall.EDQUOT) ||
		strings.Contains(strings.ToLower(err.Error()), "disk full") ||
		strings.Contains(strings.ToLower(err.Error()), "no space left")
}
