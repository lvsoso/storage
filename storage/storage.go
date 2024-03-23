package storage

import (
	"context"
	"io"
)

type StorageType string

type Opt struct {
	Key   string
	Value interface{}
}

type Signer interface {
	Sign(path string) (string, error)
}

type Storage interface {
	Get(ctx context.Context, path string,  opt ...Opt) (io.ReadCloser, error)
	Put(ctx context.Context, path string, r io.ReadCloser, size int64, opt ...Opt) (int64, error)
	Delete(ctx context.Context, path string, opt ...Opt) error
	Exist(ctx context.Context, path string) (bool, error)
}

func NewStorage(ctx context.Context, st StorageType, cfg interface{}) (Storage, error) {
	if st == "aws" {
		return newAWS(ctx, cfg)
	}
	return nil, nil
}
