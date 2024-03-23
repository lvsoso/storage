package storage

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getCfg() map[string]string {
	if os.Getenv("access_key") == "" {
		return nil
	}
	return map[string]string{
		"access_key":        os.Getenv("access_key"),
		"secret_access_key": os.Getenv("secret_key"),
		"token":             os.Getenv("token"),
		"end_point":         os.Getenv("end_point"),
		"bucket":            os.Getenv("bucket"),
	}
}

func TestPipeline(t *testing.T) {
	cfg := getCfg()
	if cfg == nil {
		t.Skip("skip test because empty env variable")
	}

	ctx := context.Background()
	st1, err := newAWS(ctx, cfg)
	assert.Nil(t, err)

	st2, err := newAWS(ctx, cfg)
	assert.Nil(t, err)

	name := RandomString(6)
	p1 := "/" + st1.cfg.Bucket + "/123/456/" + name
	p2 := "/" + st1.cfg.Bucket + "/456/123/" + name

	// PUT
	s := "abcdfsgrgrtytrytryhgdfhfghghgfh"
	buf := bytes.NewBufferString(s)

	n, err := st1.Put(ctx, p1, io.NopCloser(buf), int64(buf.Len()))
	assert.Nil(t, err)
	assert.Equal(t, n, int64(len(s)))

	// Exist
	exist, err := st1.Exist(ctx, p1)
	assert.Nil(t, err)
	assert.True(t, exist)

	// Get
	r, err := st1.Get(ctx, p1)
	assert.Nil(t, err)
	defer r.Close()

	ss, err := io.ReadAll(r)
	assert.Nil(t, err)
	assert.Equal(t, s, string(ss))

	// Copy
	r, err = st1.Get(ctx, p1)
	assert.Nil(t, err)
	defer r.Close()

	n, err = st2.Put(ctx, p2, r, int64(len(s)))
	assert.Nil(t, err)
	assert.Equal(t, n, int64(len(s)))

	r, err = st2.Get(ctx, p2)
	assert.Nil(t, err)
	defer r.Close()

	ss, err = io.ReadAll(r)
	assert.Nil(t, err)
	assert.Equal(t, s, string(ss))

	// Delete
	err = st1.Delete(ctx, p1)
	assert.Nil(t, err)

	exist, err = st1.Exist(ctx, p1)
	assert.Nil(t, err)
	assert.False(t, exist)

	err = st2.Delete(ctx, p2)
	assert.Nil(t, err)
}

func TestLargeFile(t *testing.T) {
	size := int64(70 * 1024 * 1024)
	fileName, err := GenFile(size)
	assert.Nil(t, err)
	defer func() {
		os.Remove(fileName)
	}()

	file, err := os.Open(fileName)
	assert.Nil(t, err)

	cfg := getCfg()
	if cfg == nil {
		t.Skip("skip test because empty env variable")
	}

	ctx := context.Background()
	st1, err := newAWS(ctx, cfg)
	assert.Nil(t, err)

	st2, err := newAWS(ctx, cfg)
	assert.Nil(t, err)

	name := RandomString(6)
	p1 := "/" + st1.cfg.Bucket + "/123/456/" + name
	p2 := "/" + st1.cfg.Bucket + "/456/123/" + name

	n, err := st1.putMultipart(ctx, p1, file, size)
	assert.Nil(t, err)
	assert.Equal(t, n, size)

	exist, err := st1.Exist(ctx, p1)
	assert.Nil(t, err)
	assert.True(t, exist)

	r, err := st1.Get(ctx, p1)
	assert.Nil(t, err)
	defer r.Close()

	n, err = st2.putMultipart(ctx, p2, r, size)
	assert.Nil(t, err)
	assert.Equal(t, n, size)

	exist, err = st2.Exist(ctx, p2)
	assert.Nil(t, err)
	assert.True(t, exist)

	err = st1.Delete(ctx, p1)
	assert.Nil(t, err)

	err = st2.Delete(ctx, p2)
	assert.Nil(t, err)
}
