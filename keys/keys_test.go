package keys_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ahobsonsayers/twigots/keys"
	"github.com/stretchr/testify/require"
)

func sampleRawKeys() []byte {
	return []byte(`{
		"api_key": "test-api-key",
		"User-Agent": "test-user-agent",
		"x-prosopo-site-key": "test-site-key",
		"x-prosopo-android-integrity-token": "test-integrity-token"
	}`)
}

func TestLoadKeysFromURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(sampleRawKeys())
	}))
	defer server.Close()

	k, err := keys.LoadKeysFromURL(context.Background(), server.URL)
	require.NoError(t, err)

	require.Equal(t, "test-api-key", k.APIKey())
	require.Equal(t, "test-user-agent", k.UserAgent())
	require.Equal(t, "test-site-key", k.ProsopoSiteKey())
	require.Equal(t, "test-integrity-token", k.ProsopoIntegrityToken())
	require.True(t, k.HasWatcher())
}

func TestLoadKeysFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "keys.json")
	require.NoError(t, os.WriteFile(path, sampleRawKeys(), 0o644))

	k, err := keys.LoadKeysFromFile(path)
	require.NoError(t, err)

	require.Equal(t, "test-api-key", k.APIKey())
	require.Equal(t, "test-user-agent", k.UserAgent())
	require.Equal(t, "test-site-key", k.ProsopoSiteKey())
	require.Equal(t, "test-integrity-token", k.ProsopoIntegrityToken())
	require.True(t, k.HasWatcher())
}

func TestUpdateNilLeavesUnchanged(t *testing.T) {
	k := &keys.Keys{}
	k.Update(
		ptr("api"),
		ptr("ua"),
		ptr("site"),
		ptr("token"),
	)

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

func TestGettersAndSourceChecks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "keys.json")
	require.NoError(t, os.WriteFile(path, sampleRawKeys(), 0o644))

	k, err := keys.LoadKeysFromFile(path)
	require.NoError(t, err)

	require.True(t, k.HasWatcher())
}

func TestWatchURL(t *testing.T) {
	var counter atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		counter.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(sampleRawKeys())
	}))
	defer server.Close()

	k, err := keys.LoadKeysFromURL(context.Background(), server.URL, keys.WithURLWatcherPollInterval(50*time.Millisecond))
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
	dir := t.TempDir()
	path := filepath.Join(dir, "keys.json")
	require.NoError(t, os.WriteFile(path, sampleRawKeys(), 0o644))

	k, err := keys.LoadKeysFromFile(path)
	require.NoError(t, err)

	err = k.StartWatching()
	require.NoError(t, err)
	defer k.StopWatching()

	updated := []byte(`{
		"api_key": "updated-api-key",
		"User-Agent": "updated-user-agent",
		"x-prosopo-site-key": "updated-site-key",
		"x-prosopo-android-integrity-token": "updated-integrity-token"
	}`)
	require.Eventually(t, func() bool {
		return os.WriteFile(path, updated, 0o644) == nil
	}, 2*time.Second, 20*time.Millisecond)

	require.Eventually(t, func() bool {
		return k.APIKey() == "updated-api-key"
	}, 5*time.Second, 20*time.Millisecond)

	require.Equal(t, "updated-user-agent", k.UserAgent())
	require.Equal(t, "updated-site-key", k.ProsopoSiteKey())
	require.Equal(t, "updated-integrity-token", k.ProsopoIntegrityToken())
}

func TestWatchWithoutSourceErrors(t *testing.T) {
	k := &keys.Keys{}
	err := k.StartWatching()
	require.Error(t, err)
}

func ptr(s string) *string {
	return &s
}
