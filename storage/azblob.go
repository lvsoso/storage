package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
)

var _ Storage = (*Azblob)(nil)

type Azblob struct {
	cfg    AzblobConfig
	client *azblob.Client
}

type AzblobConfig struct {
	Account          string
	Key              string
	Endpoint         string
	Container        string
	ConnectionString string
}

// newAzblob creates a storage backend of azblob.
//
// SDK: https://github.com/Azure/azure-sdk-for-go/tree/main/sdk/storage/azblob#readme
//
// Config:
// example 1:
//
//	{
//	    "account": "devstoreaccount1",
//	    "key": "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==",
//	    "endpoint": "http://127.0.0.1:10000",
//	    "conatinner": "local-test1"
//	}
//
// example 2:
//
//	{
//	    "connection_string":"DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;",
//	    "conatinner": "abc"
//	}
func newAzblob(cfg interface{}) (*Azblob, error) {
	config, ok := cfg.(map[string]string)
	if !ok {
		return nil, errInvalidConfig
	}

	azConfig := &AzblobConfig{}
	azConfig.Account = config["account"]
	azConfig.Key = config["key"]
	azConfig.Endpoint = config["endpoint"]
	azConfig.Container = config["conatinner"]
	azConfig.ConnectionString = config["connection_string"]

	if azConfig.Container == "" {
		return nil, errInvalidConfig
	}

	if azConfig.ConnectionString == "" && (azConfig.Account == "" || azConfig.Key == "" || azConfig.Endpoint == "") {
		return nil, errInvalidConfig
	}

	var client *azblob.Client
	var err error
	if azConfig.ConnectionString != "" {
		client, err = azblob.NewClientFromConnectionString(azConfig.ConnectionString, nil)
		if err != nil {
			return nil, err
		}
	} else {
		creds, err := azblob.NewSharedKeyCredential(
			azConfig.Account, azConfig.Key)
		if err != nil {
			return nil, err
		}

		t := strings.Split(azConfig.Endpoint, "://")
		schema := t[0]
		path := t[1]
		client, err = azblob.NewClientWithSharedKeyCredential(
			// "https://MYSTORAGEACCOUNT.blob.core.windows.net/"
			fmt.Sprintf("%s://%s.%s", schema, azConfig.Account, path),
			creds,
			nil,
		)
	}

	return &Azblob{
		cfg:    *azConfig,
		client: client,
	}, nil
}

func (az *Azblob) Get(ctx context.Context, path string, opt ...Opt) (io.ReadCloser, error) {
	key := az.getKeyPath(path)
	resp, err := az.client.DownloadStream(ctx,
		az.cfg.Container,
		key,
		nil,
	)
	if err != nil {
		defer resp.Body.Close()
		return nil, err
	}
	return resp.Body, nil
}

func (az *Azblob) Put(ctx context.Context, path string, r io.ReadCloser, size int64, opt ...Opt) (int64, error) {
	key := az.getKeyPath(path)
	_, err := az.client.UploadStream(ctx,
		az.cfg.Container,
		key,
		r,
		nil)
	if err != nil {
		return 0, err
	}
	return size, nil
}

func (az *Azblob) Delete(ctx context.Context, path string, opt ...Opt) error {
	key := az.getKeyPath(path)
	_, err := az.client.DeleteBlob(ctx, az.cfg.Container, key, nil)
	return err
}

func (az *Azblob) Exist(ctx context.Context, path string) (bool, error) {
	key := az.getKeyPath(path)
	_, err := az.client.ServiceClient().NewContainerClient(az.cfg.Container).
		NewBlobClient(key).GetProperties(ctx, nil)
	if err != nil {
		if bloberror.HasCode(err, bloberror.BlobNotFound) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func (az *Azblob) getKeyPath(path string) string {
	if strings.HasPrefix("/"+az.cfg.Container, path) {
		return path[len("/"+az.cfg.Container+"/")+1:]
	}
	return path[len(az.cfg.Container+"/")+1:]
}
