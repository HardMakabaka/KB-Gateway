package api

import "sync"

type KeyedMutex struct {
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

func (k *KeyedMutex) Lock(key string) func() {
	k.mu.Lock()
	if k.locks == nil {
		k.locks = make(map[string]*sync.Mutex)
	}
	m, ok := k.locks[key]
	if !ok {
		m = &sync.Mutex{}
		k.locks[key] = m
	}
	k.mu.Unlock()

	m.Lock()
	return func() { m.Unlock() }
}
