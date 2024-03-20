package main

import (
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	AddDiskConfig("local", DiskConfig{
		Driver:  "local",
		Root:    "/home/lv/lvsoso/storage/rc-storage/test_data",
		Timeout: 360 * time.Second,
	})

	// azureblob
	AddDiskConfig("azureblob", DiskConfig{
		Driver: "azureblob",
		Root:   "devstoreaccount1",
		BackendConfig: map[string]string{
			"account":            "devstoreaccount1",
			"key":                "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==",
			"endpoint":           "http://127.0.0.1:10000/devstoreaccount1",
			"chunk_size":         "4096",
			"upload_concurrency": "2",
		},
		Timeout: 360 * time.Second,
	})

	// minio
	AddDiskConfig("minio", DiskConfig{
		Driver: "s3",
		Root:   "/test1",
		BackendConfig: map[string]string{
			"type":              "s3",
			"region":            "other-v2-signature",
			"access_key_id":     "minioadmin",
			"secret_access_key": "minioadmin",
			"provider":          "Minio",
			"endpoint":          "http://127.0.0.1:9000",
			"env_auth":          "false",
			"chunk_size":        "8192",
			"no_check_bucket":   "true",
			"upload_cutoff":     "65536",
		},
		Timeout: 360 * time.Second,
	})
}

func TestStorage_GetFromS3(t *testing.T) {
	storage, err := Disk("minio")
	assert.NoError(t, err)

	_, err = storage.Get("1.png")
	assert.NotNil(t, err)
}

func TestStorage_GetFromblob(t *testing.T) {
	storage, err := Disk("azureblob")
	assert.NoError(t, err)

	_, err = storage.Get("1.txt")
	assert.NotNil(t, err)
}

func TestStorage_BlobPutGetAndDelete(t *testing.T) {
	storage, err := Disk("azureblob")
	assert.NoError(t, err)

	fileName := "BlobPutGetAndDelete"
	content := "123456789"

	file, err := storage.put(fileName, ioutil.NopCloser(strings.NewReader(content)))
	assert.NoError(t, err)
	assert.Equal(t, int64(len(content)), file.Size())

	r, err := storage.Get(fileName)

	if assert.NoError(t, err) {
		reader, err := r.Open()
		assert.NoError(t, err)

		data, err := ioutil.ReadAll(reader)
		assert.NoError(t, err)
		defer reader.Close()

		t.Log(data)
	}

	err = storage.Delete(fileName)
	assert.NoError(t, err)
	assert.False(t, storage.Exists(fileName))
}

func TestStorage_Copy(t *testing.T) {
	localStorage, err := Disk("local")
	assert.NoError(t, err)

	azureblobStorage, err := Disk("azureblob")
	assert.NoError(t, err)

	// copy
	fileName := "copytest.txt"
	content := "123456789"

	r, err := azureblobStorage.Get(fileName)
	assert.NoError(t, err)
	reader, err := r.Open()
	assert.NoError(t, err)

	file, err := localStorage.put(fileName, reader)
	assert.NoError(t, err)
	assert.Equal(t, int64(len(content)), file.Size())

	r, err = localStorage.Get(fileName)

	if assert.NoError(t, err) {
		reader, err := r.Open()
		assert.NoError(t, err)

		data, err := ioutil.ReadAll(reader)
		assert.NoError(t, err)
		defer reader.Close()

		assert.Equal(t, content, string(data))
	}
}

func TestStorage_Copy2(t *testing.T) {
	minioStorage, err := Disk("minio")
	assert.NoError(t, err)

	azureblobStorage, err := Disk("azureblob")
	assert.NoError(t, err)

	// copy
	fileName := "copytest.txt"
	content := "123456789"

	r, err := azureblobStorage.Get(fileName)
	assert.NoError(t, err)
	reader, err := r.Open()
	assert.NoError(t, err)

	file, err := minioStorage.put("/test1/"+fileName, reader)
	assert.NoError(t, err)
	assert.Equal(t, int64(len(content)), file.Size())

	r, err = minioStorage.Get(fileName)

	if assert.NoError(t, err) {
		reader, err := r.Open()
		assert.NoError(t, err)

		data, err := ioutil.ReadAll(reader)
		assert.NoError(t, err)
		defer reader.Close()

		assert.Equal(t, content, string(data))
	}
}

func TestStorage_Get(t *testing.T) {
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
