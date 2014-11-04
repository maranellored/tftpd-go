package main

import "sync"

type TftpCache struct {
	cachedFiles map[string][]byte
	accessLock  *sync.RWMutex
}

func NewTftpCache() *TftpCache {
	return &TftpCache{
		cachedFiles: make(map[string][]byte),
		accessLock:  new(sync.RWMutex),
	}

}

func (cache *TftpCache) putData(filename string, data []byte) {
	cache.accessLock.Lock()
	defer cache.accessLock.Unlock()

	cache.cachedFiles[filename] = data
}

func (cache *TftpCache) getData(filename string) []byte {
	cache.accessLock.RLock()
	defer cache.accessLock.RUnlock()

	data := cache.cachedFiles[filename]

	return data
}
