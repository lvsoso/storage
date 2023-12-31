package main

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	AddDiskConfig("local", DiskConfig{
		Driver: "local",
		Root:   "/home/lv/lvsoso/storage/rc-storage/test_data",
	})
}

func Test_get(t *testing.T) {
	storage, err := Disk("local")
	assert.NoError(t, err)

	r, err := storage.Get("1.txt")

	if assert.NoError(t, err) {
		reader, err := r.Open()
		assert.NoError(t, err)

		data, err := ioutil.ReadAll(reader)
		assert.NoError(t, err)
		defer reader.Close()

		t.Log(data)
	}
}

func TestStorage_Exists(t *testing.T) {
	storage, err := Disk("local")
	assert.NoError(t, err)

	t.Run("file exist", func(t *testing.T) {
		ok := storage.Exists("2/2.txt")
		assert.True(t, ok, "json file is missing it can't be")
	})

	t.Run("file is missing", func(t *testing.T) {
		ok := storage.Exists("test-missing.json")
		assert.False(t, ok, "json file is exists it can't be")
	})
}

func TestStorage_PutAndDelete(t *testing.T) {
	storage, err := Disk("local")
	assert.NoError(t, err)

	content := "123456789"

	file, err := storage.put("test2.json", ioutil.NopCloser(strings.NewReader(content)))
	assert.NoError(t, err)
	assert.Equal(t, int64(len(content)), file.Size())

	err = storage.Delete("test2.json")
	assert.NoError(t, err)
	assert.False(t, storage.Exists("test2.json"))
}
