package keys

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const fileDebounce = 200 * time.Millisecond

// FileWatcher watches a file for key updates.
type FileWatcher struct {
	path string

	watcher *fsnotify.Watcher

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

type FileWatcherOpt func(*FileWatcher)

func WithFileWatcherCtx(ctx context.Context) FileWatcherOpt {
	return func(w *FileWatcher) {
		w.ctx = ctx
	}
}

func NewFileWatcher(path string, opts ...FileWatcherOpt) *FileWatcher {
	// Apply options
	watcher := &FileWatcher{path: path}
	for _, opt := range opts {
		opt(watcher)
	}

	// Set defaults
	if watcher.ctx == nil {
		watcher.ctx = context.Background()
	}

	return watcher
}

func (w *FileWatcher) StartWatching(keys *Keys) error {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	fileDirectory := filepath.Dir(w.path)
	fileName := filepath.Base(w.path)

	err = fsWatcher.Add(fileDirectory)
	if err != nil {
		fsWatcher.Close()
		return fmt.Errorf("failed to watch directory %q: %w", fileDirectory, err)
	}

	w.watcher = fsWatcher

	ctx, cancel := context.WithCancel(w.ctx)
	w.cancel = cancel
	w.wg.Add(1)

	go w.watch(ctx, keys, fileName)

	return nil
}

func (w *FileWatcher) StopWatching() {
	if w.cancel != nil {
		w.cancel()
	}
	w.wg.Wait()
	if w.watcher != nil {
		w.watcher.Close()
	}
}

func (w *FileWatcher) watch(ctx context.Context, keys *Keys, targetFileName string) {
	defer w.wg.Done()

	var debounce *time.Timer
	for {
		select {
		case <-ctx.Done():
			return

		case event, ok := <-w.watcher.Events:
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
					err := w.load(keys)
					if err != nil {
						slog.Error(
							"failed to reload keys",
							"path", w.path,
							"error", err,
						)
					}
				},
			)

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}

			slog.Error(
				"keys file watcher error",
				"error", err,
			)
		}
	}
}

func (w *FileWatcher) load(keys *Keys) error {
	keysJson, err := os.ReadFile(w.path)
	if err != nil {
		return fmt.Errorf("failed to read keys file: %w", err)
	}

	return keys.UpdateFromJSON(keysJson)
}

func LoadKeysFromFile(path string, watcherOpts ...FileWatcherOpt) (*Keys, error) {
	watcher := NewFileWatcher(path, watcherOpts...)

	keys := &Keys{}
	err := watcher.load(keys)
	if err != nil {
		return nil, err
	}

	keys.SetWatcher(watcher)
	return keys, nil
}
