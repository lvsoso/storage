package storage

import (
	"math/rand"
	"os"
	"time"
	"unsafe"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return *(*string)(unsafe.Pointer(&b))
}

func GenFile(size int64) (string, error) {
	file, err := os.CreateTemp(os.TempDir(), "st-*")
	if err != nil {
		return "", err
	}
	defer file.Close()

	currentSize := int64(0)
	buf := make([]byte, 1024)

	for currentSize < size {
		if _, err := rand.Read(buf); err != nil {
			return "", err
		}
		n, err := file.Write(buf)
		if err != nil {
			return "", err
		}
		currentSize += int64(n)
	}

	remaining := size - currentSize
	if remaining > 0 {
		buf = buf[:remaining]
		if _, err := rand.Read(buf); err != nil {
			return "", err
		}
		if _, err := file.Write(buf); err != nil {
			return "", err
		}
	}

	if err := file.Truncate(size); err != nil {
		return "", err
	}

	if err := file.Sync(); err != nil {
		return "", err
	}

	return file.Name(), nil
}
