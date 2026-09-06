package keys

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
)

const fileDebounce = 200 * time.Millisecond

// FileSource uses a file as a key source.
// Watching the source will watch the file for updates.
type FileSource struct {
	path string

	source *fsnotify.Watcher

	watching atomic.Bool

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

type FileSourceOpt func(*FileSource)

func WithFileSourceCtx(ctx context.Context) FileSourceOpt {
	return func(w *FileSource) {
		w.ctx = ctx
	}
}

func NewFileSource(path string, opts ...FileSourceOpt) *FileSource {
	// Apply options
	source := &FileSource{path: path}
	for _, opt := range opts {
		opt(source)
	}

	// Set defaults
	if source.ctx == nil {
		source.ctx = context.Background()
	}

	return source
}

func (w *FileSource) StartWatching(keys *Keys) error {
	fsnotifyWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file source: %w", err)
	}

	fileDirectory := filepath.Dir(w.path)
	fileName := filepath.Base(w.path)

	err = fsnotifyWatcher.Add(fileDirectory)
	if err != nil {
		fsnotifyWatcher.Close()
		return fmt.Errorf("failed to watch directory %q: %w", fileDirectory, err)
	}

	w.source = fsnotifyWatcher

	ctx, cancel := context.WithCancel(w.ctx)
	w.cancel = cancel
	w.wg.Add(1)

	w.watching.Store(true)

	go w.watch(ctx, keys, fileName)

	return nil
}

func (w *FileSource) StopWatching() {
	if w.cancel != nil {
		w.cancel()
	}
	w.wg.Wait()
	if w.source != nil {
		w.source.Close()
	}

	w.watching.Store(false)
}

func (w *FileSource) IsWatching() bool {
	return w.watching.Load()
}

func (w *FileSource) watch(ctx context.Context, keys *Keys, targetFileName string) {
	defer w.wg.Done()

	var debounce *time.Timer
	for {
		select {
		case <-ctx.Done():
			return

		case event, ok := <-w.source.Events:
			if !ok {
				return
			}

			fileName := filepath.Base(event.Name)
			if fileName != targetFileName {
				continue
			}

			if debounce != nil {
				debounce.Stop()
			}

			debounce = time.AfterFunc(
				fileDebounce,
				func() {
					err := w.Load(keys)
					if err != nil {
						slog.Error(
							"failed to reload keys",
							"path", w.path,
							"error", err,
						)
					}
				},
			)

		case err, ok := <-w.source.Errors:
			if !ok {
				return
			}

			slog.Error(
				"keys file source error",
				"error", err,
			)
		}
	}
}

func (w *FileSource) Load(keys *Keys) error {
	keysJson, err := os.ReadFile(w.path)
	if err != nil {
		return fmt.Errorf("failed to load keys: %w", err)
	}

	return keys.UpdateFromJSON(keysJson)
}

func FromFile(path string, sourceOpts ...FileSourceOpt) (*Keys, error) {
	keys := &Keys{}
	err := keys.SetSource(NewFileSource(path, sourceOpts...))
	if err != nil {
		return nil, err
	}

	return keys, nil
}
