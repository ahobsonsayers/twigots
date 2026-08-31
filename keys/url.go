package keys

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

const defaultURLPollInterval = time.Minute

// URLWatcher watches a URL for key updates by polling at a regular interval.
type URLWatcher struct {
	client       *http.Client
	url          string
	pollInterval time.Duration

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

type URLWatcherOpt func(*URLWatcher)

func WithURLWatcherCtx(ctx context.Context) URLWatcherOpt {
	return func(w *URLWatcher) {
		w.ctx = ctx
	}
}

func WithURLWatcherHTTPClient(client *http.Client) URLWatcherOpt {
	return func(w *URLWatcher) {
		w.client = client
	}
}

func WithURLWatcherPollInterval(d time.Duration) URLWatcherOpt {
	return func(w *URLWatcher) {
		w.pollInterval = d
	}
}

func NewURLWatcher(url string, opts ...URLWatcherOpt) *URLWatcher {
	// Apply options
	watcher := &URLWatcher{url: url}
	for _, opt := range opts {
		opt(watcher)
	}

	// Set defaults
	if watcher.client == nil {
		watcher.client = http.DefaultClient
	}
	if watcher.pollInterval <= 0 {
		watcher.pollInterval = defaultURLPollInterval
	}
	if watcher.ctx == nil {
		watcher.ctx = context.Background()
	}

	return watcher
}

func (w *URLWatcher) StartWatching(keys *Keys) error {
	ctx, cancel := context.WithCancel(w.ctx)
	w.cancel = cancel
	w.wg.Add(1)

	go w.watch(ctx, keys)

	return nil
}

func (w *URLWatcher) StopWatching() {
	if w.cancel != nil {
		w.cancel()
	}
	w.wg.Wait()
}

func (w *URLWatcher) watch(ctx context.Context, keys *Keys) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			err := w.fetch(ctx, keys)
			if err != nil {
				slog.Error(
					"failed to fetch keys",
					"error", err,
				)
			}
		}
	}
}

func (w *URLWatcher) fetch(ctx context.Context, keys *Keys) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, w.url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	response, err := w.client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to fetch keys: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status fetching keys: %s", response.Status)
	}

	keysJson, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed reading keys response: %w", err)
	}

	return keys.UpdateFromJSON(keysJson)
}

func LoadKeysFromURL(ctx context.Context, url string, watcherOpts ...URLWatcherOpt) (*Keys, error) {
	watcher := NewURLWatcher(url, watcherOpts...)

	keys := &Keys{}
	err := watcher.fetch(ctx, keys)
	if err != nil {
		return nil, err
	}

	keys.SetWatcher(watcher)
	return keys, nil
}
