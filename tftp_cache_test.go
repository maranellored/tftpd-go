package main

import (
	"bytes"
	"testing"
)

func TestCache(t *testing.T) {
	cache := NewTftpCache()

	data := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'}
	key := "alphabet"

	cache.putData(key, data)

	value := cache.getData(key)

	if !bytes.Equal(value, data) {
		t.Error("Expected to get the same value after put")
	}
}

func TestCacheFail(t *testing.T) {
	cache := NewTftpCache()
	key := "string"

	value := cache.getData(key)
	if value != nil {
		t.Error("Expected no data for non-existent key")
	}
}
