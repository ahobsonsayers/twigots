package keys_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ahobsonsayers/twigots/keys"
	"github.com/stretchr/testify/require"
)

func rawKeys(prefix string) []byte {
	return []byte(`{
		"api_key": "` + prefix + `-api-key",
		"User-Agent": "` + prefix + `-user-agent",
		"x-prosopo-site-key": "` + prefix + `-site-key",
		"x-prosopo-android-integrity-token": "` + prefix + `-integrity-token"
	}`)
}

func requireKeys(t *testing.T, k *keys.Keys, prefix string) {
	t.Helper()
	require.Equal(t, prefix+"-api-key", k.APIKey())
	require.Equal(t, prefix+"-user-agent", k.UserAgent())
	require.Equal(t, prefix+"-site-key", k.ProsopoSiteKey())
	require.Equal(t, prefix+"-integrity-token", k.ProsopoIntegrityToken())
}

func TestFromURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(rawKeys("test"))
	}))
	defer server.Close()

	k, err := keys.FromURL(server.URL)
	require.NoError(t, err)

	requireKeys(t, k, "test")
	require.True(t, k.HasSource())
}

func TestFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "keys.json")
	require.NoError(t, os.WriteFile(path, rawKeys("test"), 0o644))

	k, err := keys.FromFile(path)
	require.NoError(t, err)

	requireKeys(t, k, "test")
	require.True(t, k.HasSource())
}

func TestUpdate(t *testing.T) {
	k := &keys.Keys{}
	k.Update(ptr("api"), ptr("ua"), ptr("site"), ptr("token"))

	require.Equal(t, "api", k.APIKey())
	require.Equal(t, "ua", k.UserAgent())
	require.Equal(t, "site", k.ProsopoSiteKey())
	require.Equal(t, "token", k.ProsopoIntegrityToken())

	k.Update(nil, nil, nil, nil)

	require.Equal(t, "api", k.APIKey())
	require.Equal(t, "ua", k.UserAgent())
	require.Equal(t, "site", k.ProsopoSiteKey())
	require.Equal(t, "token", k.ProsopoIntegrityToken())

	k.Update(ptr("api2"), nil, ptr("site2"), nil)

	require.Equal(t, "api2", k.APIKey())
	require.Equal(t, "ua", k.UserAgent())
	require.Equal(t, "site2", k.ProsopoSiteKey())
	require.Equal(t, "token", k.ProsopoIntegrityToken())
}

func TestWatchURL(t *testing.T) {
	var counter atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		counter.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(rawKeys("test"))
	}))
	defer server.Close()

	k, err := keys.FromURL(server.URL, keys.WithURLSourcePollInterval(50*time.Millisecond))
	require.NoError(t, err)

	err = k.StartWatching()
	require.NoError(t, err)
	defer k.StopWatching()

	require.Eventually(t, func() bool {
		return counter.Load() > 1
	}, 5*time.Second, 20*time.Millisecond)

	require.Equal(t, "test-api-key", k.APIKey())
}

func TestWatchFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "keys.json")
	require.NoError(t, os.WriteFile(path, rawKeys("test"), 0o644))

	k, err := keys.FromFile(path)
	require.NoError(t, err)

	err = k.StartWatching()
	require.NoError(t, err)
	defer k.StopWatching()

	require.NoError(t, os.WriteFile(path, rawKeys("updated"), 0o644))

	require.Eventually(t, func() bool {
		return k.APIKey() == "updated-api-key"
	}, 5*time.Second, 20*time.Millisecond)

	requireKeys(t, k, "updated")
}

type fakeSource struct {
	mu       sync.Mutex
	watching bool
}

func (*fakeSource) Load(_ *keys.Keys) error {
	return nil
}

func (f *fakeSource) StartWatching(_ *keys.Keys) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.watching = true
	return nil
}

func (f *fakeSource) StopWatching() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.watching = false
}

func (f *fakeSource) IsWatching() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.watching
}

func TestKeysLoadFromFileSource(t *testing.T) {
	path := filepath.Join(t.TempDir(), "keys.json")
	require.NoError(t, os.WriteFile(path, rawKeys("test"), 0o644))

	k, err := keys.FromFile(path)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(path, rawKeys("updated"), 0o644))
	require.NoError(t, k.Load())

	requireKeys(t, k, "updated")
}

func TestKeysLoadWithoutSourceErrors(t *testing.T) {
	k := &keys.Keys{}
	err := k.Load()
	require.Error(t, err)
}

func TestSetSource(t *testing.T) {
	var counter atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		counter.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(rawKeys("test"))
	}))
	defer server.Close()

	k := &keys.Keys{}
	k.Update(ptr("old-api-key"), ptr("old-user-agent"), ptr("old-site-key"), ptr("old-integrity-token"))

	oldSource := &fakeSource{}
	require.NoError(t, k.SetSource(oldSource))
	require.NoError(t, k.StartWatching())

	newSource := keys.NewURLSource(server.URL, keys.WithURLSourcePollInterval(50*time.Millisecond))
	err := k.SetSource(newSource)
	require.NoError(t, err)
	defer k.StopWatching()

	require.False(t, oldSource.IsWatching(), "old source should be stopped")

	requireKeys(t, k, "test")

	require.Eventually(t, func() bool {
		return counter.Load() > 1
	}, 5*time.Second, 20*time.Millisecond)
}

func TestSetSourceLoadErrorLeavesUnchanged(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	k := &keys.Keys{}
	k.Update(ptr("api"), ptr("ua"), ptr("site"), ptr("token"))

	oldSource := &fakeSource{}
	require.NoError(t, k.SetSource(oldSource))
	require.NoError(t, k.StartWatching())

	err := k.SetSource(keys.NewURLSource(server.URL))
	require.Error(t, err)

	require.Equal(t, "api", k.APIKey())
	require.True(t, oldSource.IsWatching(), "old source should be untouched")
	require.True(t, k.HasSource())
	require.True(t, k.IsWatching())
}

func TestSetSourceFileLoadErrorLeavesUnchanged(t *testing.T) {
	k := &keys.Keys{}
	k.Update(ptr("api"), nil, nil, nil)

	err := k.SetSource(keys.NewFileSource(filepath.Join(t.TempDir(), "missing.json")))
	require.Error(t, err)

	require.Equal(t, "api", k.APIKey())
	require.False(t, k.HasSource())
}

func TestSetSourceNotWatchingStaysNotWatching(t *testing.T) {
	var counter atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		counter.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(rawKeys("test"))
	}))
	defer server.Close()

	k, err := keys.FromURL(server.URL)
	require.NoError(t, err)

	err = k.SetSource(keys.NewURLSource(server.URL, keys.WithURLSourcePollInterval(50*time.Millisecond)))
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)
	require.Equal(t, int32(2), counter.Load(), "source should not have started polling")
}

func TestWatchWithoutSourceErrors(t *testing.T) {
	k := &keys.Keys{}
	err := k.StartWatching()
	require.Error(t, err)
}

func ptr(s string) *string {
	return &s
}
