package keys

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

type Keys struct {
	mu sync.RWMutex

	apiKey                string
	userAgent             string
	prosopoSiteKey        string
	prosopoIntegrityToken string

	source Source
}

// Source is a keys source.
// Must be able to watch source for updates
type Source interface {
	Load(k *Keys) error
	StartWatching(k *Keys) error
	StopWatching()
	IsWatching() bool
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
	var keys struct {
		APIKey                string `json:"api_key"`
		UserAgent             string `json:"User-Agent"`
		ProsopoSiteKey        string `json:"x-prosopo-site-key"`
		ProsopoIntegrityToken string `json:"x-prosopo-android-integrity-token"`
	}

	err := json.Unmarshal(keysJson, &keys)
	if err != nil {
		return fmt.Errorf("failed to unmarshal keys: %w", err)
	}

	k.Update(
		&keys.APIKey,
		&keys.UserAgent,
		&keys.ProsopoSiteKey,
		&keys.ProsopoIntegrityToken,
	)

	return nil
}

func (k *Keys) HasSource() bool {
	return k.source != nil
}

// Load reloads the keys from the current source.
func (k *Keys) Load() error {
	if k.source == nil {
		return errors.New("no source configured")
	}
	return k.source.Load(k)
}

func (k *Keys) StartWatching() error {
	if k.source == nil {
		return errors.New("no source configured")
	}
	return k.source.StartWatching(k)
}

func (k *Keys) StopWatching() {
	if k.source != nil && k.source.IsWatching() {
		k.source.StopWatching()
	}
}

func (k *Keys) IsWatching() bool {
	if k.source == nil {
		return false
	}
	return k.source.IsWatching()
}

// SetSource replaces the keys source.
// Keys are loaded from the new source immediately, and if watch is currently
// running, the new source will start being watched.
// If load fails, no changes to keys are made.
func (k *Keys) SetSource(source Source) error {
	err := source.Load(k)
	if err != nil {
		return fmt.Errorf("failed to load keys from new source: %w", err)
	}

	wasWatching := k.source != nil && k.source.IsWatching()
	if wasWatching {
		k.source.StopWatching()
	}

	k.source = source

	if wasWatching {
		err := source.StartWatching(k)
		if err != nil {
			return fmt.Errorf("failed to watch new source: %w", err)
		}
	}

	return nil
}
