package constants

import "sync"

var header string

var headerMutex = &sync.RWMutex{}

// SetHeader -
func SetHeader(h string) {
	headerMutex.Lock()
	header = h
	headerMutex.Unlock()
}

// GetHeader -
func GetHeader() string {
	result := ""

	headerMutex.RLock()
	result = header
	headerMutex.RUnlock()

	return result
}
