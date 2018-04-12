package memory

import (
	"github.com/phachon/fasthttpsession"
	"sync"
	"time"
	"errors"
	"reflect"
)

// session memory provider

const ProviderName = "memory"

type Provider struct {
	lock sync.RWMutex
	config *Config
	values map[string]*Store
}

// new memory provider
func NewProvider() *Provider {
	return &Provider{
		config: &Config{},
		values: make(map[string]*Store),
	}
}

// init provider config
func (mp *Provider) Init(memoryConfig fasthttpsession.ProviderConfig) error {
	if memoryConfig.Name() != ProviderName {
		return errors.New("session memory provider init error, config must memory config")
	}
	vc := reflect.ValueOf(memoryConfig)
	mc := vc.Interface().(*Config)
	mp.config = mc
	return nil
}

// session garbage collection
func (mp *Provider) GC(sessionLifetime int64) {
	mp.lock.RLock()
	for sessionId, value := range mp.values {
		if time.Now().Unix() >= value.lastActiveTime + sessionLifetime {
			mp.lock.RUnlock()
			// destroy session sessionId
			mp.Destroy(sessionId)
			return
		}
	}
	mp.lock.RUnlock()
}

// session id is exist
func (mp *Provider) SessionIdIsExist(sessionId string) bool {
	mp.lock.RLock()
	defer mp.lock.RUnlock()
	_, ok := mp.values[sessionId]
	if ok {
		return true
	}
	return false
}

// read session store by session id
func (mp *Provider) ReadStore(sessionId string) (fasthttpsession.SessionStore, error) {
	mp.lock.RLock()
	memStore, ok := mp.values[sessionId]
	if ok {
		mp.lock.RUnlock()
		return memStore, nil
	}
	mp.lock.RUnlock()

	memStore = NewMemoryStore(sessionId)
	mp.lock.Lock()
	mp.values[sessionId] = memStore
	mp.lock.Unlock()

	return memStore, nil
}

// regenerate session
func (mp *Provider) Regenerate(oldSessionId string, sessionId string) (fasthttpsession.SessionStore, error) {
	mp.lock.RLock()
	memStore, ok := mp.values[oldSessionId]
	if ok {
		mp.lock.RUnlock()
		// insert new session store
		mp.lock.Lock()
		newMemStore := NewMemoryStoreData(sessionId, memStore.GetAll())
		mp.values[sessionId] = newMemStore
		// delete old session store
		delete(mp.values, oldSessionId)
		mp.lock.Unlock()
		return newMemStore, nil
	}
	mp.lock.RUnlock()

	memStore = NewMemoryStore(sessionId)
	mp.lock.Lock()
	mp.values[sessionId] = memStore
	mp.lock.Unlock()

	return memStore, nil
}

// destroy session by sessionId
func (mp *Provider) Destroy(sessionId string) error {
	mp.lock.Lock()
	defer mp.lock.Unlock()
	delete(mp.values, sessionId)
	return nil
}

// session values count
func (mp *Provider) Count() int {
	return len(mp.values)
}

// register session provider
func init()  {
	fasthttpsession.Register(ProviderName, NewProvider())
}