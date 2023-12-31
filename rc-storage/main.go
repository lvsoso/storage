package main

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"path"
	"rcs/fs"
	"time"

	"github.com/gabriel-vasile/mimetype"
	_ "github.com/rclone/rclone/backend/all" // import all backends
	rfs "github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/config/configmap"
	"github.com/rclone/rclone/fs/object"
	"github.com/rclone/rclone/fs/operations"
	"github.com/rs/xid"
)

type storage struct {
	backend rfs.Fs
	Config  *DiskConfig
}

func (s *storage) Get(path string) (f fs.File, err error) {
	create := func(key string) (value interface{}, ok bool, err error) {
		ctx, cancel := context.WithTimeout(context.Background(), s.Config.Timeout)
		defer cancel()

		o, err := s.backend.NewObject(ctx, key)
		if err != nil {
			return nil, false, err
		}

		return fs.ObjectWrapper(o), true, nil
	}

	value, ok, err := create(path)
	if err != nil && !ok {
		return nil, err
	}

	return value.(fs.File), nil
}

func (s *storage) Exists(path string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), s.Config.Timeout)
	defer cancel()

	if ok, _ := rfs.FileExists(ctx, s.backend, path); ok {
		return true
	}

	return false
}

func (s *storage) Put(path string, in io.ReadCloser, metadata ...*fs.HTTPOption) (fs.File, error) {
	o, err := s.put(path, in, metadata...)
	if err != nil {
		return nil, err
	}

	return fs.ObjectWrapper(o), nil
}

func (s *storage) PutFile(dir string, in io.ReadCloser, metadata ...*fs.HTTPOption) (fs.File, error) {
	body := &bytes.Buffer{}
	mime, err := mimetype.DetectReader(io.TeeReader(in, body))
	if err != nil {
		return nil, err
	}

	id := xid.New().String()
	extension := mime.Extension()

	o, err := s.put(path.Join(dir, id+extension), ioutil.NopCloser(body), metadata...)
	if err != nil {
		return nil, err
	}

	return fs.ObjectWrapper(o), nil
}

func (s *storage) put(path string, in io.ReadCloser, metadata ...*fs.HTTPOption) (rfs.Object, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.Config.Timeout)
	defer cancel()

	var options []rfs.OpenOption
	for _, option := range metadata {
		options = append(options, option)
	}

	objInfo := object.NewStaticObjectInfo(path, time.Now(), -1, false, nil, nil)
	o, err := s.backend.Put(ctx, in, objInfo, options...)
	if err != nil {
		return nil, err
	}

	return o, nil
}

func (s *storage) Delete(paths ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.Config.Timeout)
	defer cancel()

	delChan := make(rfs.ObjectsChan)
	delErr := make(chan error, 1)
	go func() {
		delErr <- operations.DeleteFiles(ctx, delChan)
	}()
	for _, p := range paths {
		if o, err := s.backend.NewObject(ctx, p); err == nil {
			delChan <- o
		}
	}
	close(delChan)

	return <-delErr
}

// Disk _
func Disk(name string) (*storage, error) {
	return newDisk(name)
}

func newDisk(name string) (*storage, error) {
	cfg, err := getDiskConfig(name)
	if err != nil {
		return nil, err
	}

	// default timeout
	if cfg.Timeout == 0 {
		cfg.Timeout = time.Second * 30
	}

	regInfo, err := rfs.Find(cfg.Driver)
	if err != nil {
		return nil, err
	}

	//cm := fs.ConfigMap(regInfo, driver)
	cm := configmap.New()
	cm.AddGetter(cfg, configmap.PriorityDefault)

	backend, err := regInfo.NewFs(context.TODO(), name, cfg.Root, cm)
	if err != nil {
		return nil, err
	}

	return &storage{
		backend: backend,
		Config:  cfg,
	}, nil
}

func main() {

}
