package keys

import (
	"encoding/json"
	"fmt"
	"sync"
)

type Keys struct {
	mu sync.RWMutex

	apiKey                string
	userAgent             string
	prosopoSiteKey        string
	prosopoIntegrityToken string

	watcher Watcher
}

// Watcher watches a keys source for updates.
type Watcher interface {
	StartWatching(k *Keys) error
	StopWatching()
}

func (k *Keys) APIKey() string {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.apiKey
}

func (k *Keys) UserAgent() string {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.userAgent
}

func (k *Keys) ProsopoSiteKey() string {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.prosopoSiteKey
}

func (k *Keys) ProsopoIntegrityToken() string {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.prosopoIntegrityToken
}

func (k *Keys) Update(apiKey, userAgent, prosopoSiteKey, prosopoIntegrityToken *string) {
	k.mu.Lock()
	defer k.mu.Unlock()
	if apiKey != nil {
		k.apiKey = *apiKey
	}
	if userAgent != nil {
		k.userAgent = *userAgent
	}
	if prosopoSiteKey != nil {
		k.prosopoSiteKey = *prosopoSiteKey
	}
	if prosopoIntegrityToken != nil {
		k.prosopoIntegrityToken = *prosopoIntegrityToken
	}
}

func (k *Keys) UpdateFromJSON(keysJson []byte) error {
	type keys struct {
		APIKey                string `json:"api_key"`
		UserAgent             string `json:"User-Agent"`
		ProsopoSiteKey        string `json:"x-prosopo-site-key"`
		ProsopoIntegrityToken string `json:"x-prosopo-android-integrity-token"`
	}

	var unmarshaledKeys keys
	err := json.Unmarshal(keysJson, &unmarshaledKeys)
	if err != nil {
		return fmt.Errorf("failed to unmarshal keys: %w", err)
	}

	k.Update(
		&unmarshaledKeys.APIKey,
		&unmarshaledKeys.UserAgent,
		&unmarshaledKeys.ProsopoSiteKey,
		&unmarshaledKeys.ProsopoIntegrityToken,
	)

	return nil
}

func (k *Keys) HasWatcher() bool {
	return k.watcher != nil
}

func (k *Keys) SetWatcher(watcher Watcher) {
	k.watcher = watcher
}

func (k *Keys) StartWatching() error {
	if k.watcher == nil {
		return fmt.Errorf("no watcher configured")
	}

	return k.watcher.StartWatching(k)
}

func (k *Keys) StopWatching() {
	if k.watcher != nil {
		k.watcher.StopWatching()
	}
}
