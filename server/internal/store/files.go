package store

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/klauspost/compress/zstd"
	"modders-feedback/server/internal/domain"
)

const (
	archiveFast = "fast"
	archiveMax  = "max"
	zstdMagic   = "\x28\xb5\x2f\xfd"
)

type archiveJob struct {
	id     string
	status string
}

// Files keeps uploaded feedback files outside the SQLite database.
// Active feedback uses a fast zstd level so a page can open within a couple
// of seconds. Resolved feedback is rewritten at the highest level.
type Files struct {
	root     string
	rootLock sync.RWMutex

	jobs chan archiveJob
	once sync.Once
	stop chan struct{}
}

func NewFiles(root string) (*Files, error) {
	files := &Files{jobs: make(chan archiveJob, 64), stop: make(chan struct{})}
	if err := files.SetRoot(root); err != nil {
		return nil, err
	}
	return files, nil
}

func (f *Files) SetRoot(root string) error {
	root = strings.TrimSpace(root)
	if root == "" {
		return errors.New("attachment storage path is empty")
	}
	if !filepath.IsAbs(root) {
		return errors.New("attachment storage path must be absolute")
	}
	if err := os.MkdirAll(filepath.Join(root, "objects"), 0o750); err != nil {
		return err
	}
	f.rootLock.Lock()
	f.root = root
	f.rootLock.Unlock()
	return nil
}

func (f *Files) Root() string {
	f.rootLock.RLock()
	defer f.rootLock.RUnlock()
	return f.root
}

func (f *Files) Start() {
	f.once.Do(func() {
		go f.loop()
	})
}

func (f *Files) Close() {
	select {
	case <-f.stop:
	default:
		close(f.stop)
	}
}

func (f *Files) loop() {
	for {
		select {
		case <-f.stop:
			return
		case job := <-f.jobs:
			_ = f.recompress(job.id, job.status)
		}
	}
}

func (f *Files) Schedule(id, status string) {
	if !validObjectID(id) {
		return
	}
	select {
	case f.jobs <- archiveJob{id: id, status: status}:
	default:
		go func() { _ = f.recompress(id, status) }()
	}
}

func (f *Files) Save(id string, raw []byte, status string) error {
	if !validObjectID(id) {
		return errors.New("invalid attachment id")
	}
	packed, err := pack(raw, status)
	if err != nil {
		return err
	}
	return writeAtomic(f.objectPath(id), packed)
}

func (f *Files) Open(id string) ([]byte, error) {
	if !validObjectID(id) {
		return nil, os.ErrNotExist
	}
	packed, err := os.ReadFile(f.objectPath(id))
	if err != nil {
		return nil, err
	}
	return unpack(packed)
}

func (f *Files) Delete(id string) error {
	if !validObjectID(id) {
		return nil
	}
	err := os.Remove(f.objectPath(id))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (f *Files) recompress(id, status string) error {
	raw, err := f.Open(id)
	if err != nil {
		return err
	}
	return f.Save(id, raw, status)
}

func (f *Files) objectPath(id string) string {
	return filepath.Join(f.Root(), "objects", id[:2], id)
}

func pack(raw []byte, status string) ([]byte, error) {
	level := zstd.SpeedDefault
	label := archiveFast
	if status == domain.StatusResolved {
		level = zstd.SpeedBestCompression
		label = archiveMax
	}
	encoder, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(level), zstd.WithEncoderConcurrency(1))
	if err != nil {
		return nil, err
	}
	defer encoder.Close()
	frame := encoder.EncodeAll(raw, nil)
	return append([]byte(label+"\n"), frame...), nil
}

func unpack(packed []byte) ([]byte, error) {
	frame := packed
	if bytes.HasPrefix(packed, []byte(archiveFast+"\n")) || bytes.HasPrefix(packed, []byte(archiveMax+"\n")) {
		_, rest, _ := bytes.Cut(packed, []byte("\n"))
		frame = rest
	}
	if !bytes.HasPrefix(frame, []byte(zstdMagic)) {
		return nil, errors.New("attachment is not a zstd frame")
	}
	decoder, err := zstd.NewReader(nil, zstd.WithDecoderConcurrency(1))
	if err != nil {
		return nil, err
	}
	defer decoder.Close()
	return decoder.DecodeAll(frame, nil)
}

func writeAtomic(path string, payload []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".upload-*")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	ok := false
	defer func() {
		if !ok {
			_ = os.Remove(tempName)
		}
	}()
	if _, err = temp.Write(payload); err != nil {
		_ = temp.Close()
		return err
	}
	if err = temp.Close(); err != nil {
		return err
	}
	if err = os.Rename(tempName, path); err != nil {
		return err
	}
	ok = true
	return nil
}

func validObjectID(id string) bool {
	if len(id) != 32 {
		return false
	}
	for i := range len(id) {
		c := id[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}
