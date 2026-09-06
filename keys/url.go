package keys

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const defaultURLPollInterval = time.Minute

// URLSource uses a URL as a key source.
// Watching the source will poll the at a regular interval.
type URLSource struct {
	client       *http.Client
	url          string
	pollInterval time.Duration

	watching atomic.Bool

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

type URLSourceOpt func(*URLSource)

func WithURLSourceCtx(ctx context.Context) URLSourceOpt {
	return func(w *URLSource) {
		w.ctx = ctx
	}
}

func WithURLSourceHTTPClient(client *http.Client) URLSourceOpt {
	return func(w *URLSource) {
		w.client = client
	}
}

func WithURLSourcePollInterval(d time.Duration) URLSourceOpt {
	return func(w *URLSource) {
		w.pollInterval = d
	}
}

func NewURLSource(url string, opts ...URLSourceOpt) *URLSource {
	// Apply options
	source := &URLSource{url: url}
	for _, opt := range opts {
		opt(source)
	}

	// Set defaults
	if source.client == nil {
		source.client = http.DefaultClient
	}
	if source.pollInterval <= 0 {
		source.pollInterval = defaultURLPollInterval
	}
	if source.ctx == nil {
		source.ctx = context.Background()
	}

	return source
}

func (w *URLSource) StartWatching(keys *Keys) error {
	ctx, cancel := context.WithCancel(w.ctx)
	w.cancel = cancel
	w.wg.Add(1)

	w.watching.Store(true)

	go w.watch(ctx, keys)

	return nil
}

func (w *URLSource) StopWatching() {
	w.cancel()
	w.wg.Wait()

	w.watching.Store(false)
}

func (w *URLSource) IsWatching() bool {
	return w.watching.Load()
}

// Load loads the current keys from the URL and updates the keys.
func (w *URLSource) Load(keys *Keys) error {
	return w.fetch(w.ctx, keys)
}

func (w *URLSource) watch(ctx context.Context, keys *Keys) {
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
					"failed to load keys",
					"error", err,
				)
			}
		}
	}
}

func (w *URLSource) fetch(ctx context.Context, keys *Keys) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, w.url, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	response, err := w.client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to fetch url: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status fetching url: %s", response.Status)
	}

	keysJson, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed reading keys response: %w", err)
	}

	return keys.UpdateFromJSON(keysJson)
}

func FromURL(url string, sourceOpts ...URLSourceOpt) (*Keys, error) {
	keys := &Keys{}
	err := keys.SetSource(NewURLSource(url, sourceOpts...))
	if err != nil {
		return nil, err
	}

	return keys, nil
}
