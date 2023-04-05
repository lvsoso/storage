package main

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	MaxChunk                         = 10000
	ChunkSize                        = 16 * 1024 * 1024
	MaxBlobSize                      = MaxChunk * ChunkSize
	PresignedUploadPartUrlExpireTime = time.Hour * 1
)

type OssConf struct {
	AliasedURL      string
	AccessKeyID     string
	AecretAccessKey string
	Bucket          string
	BasePath        string
	Location        string
	Secure          bool
}

type OssClient struct {
	conf       OssConf
	client     *minio.Client
	coreClient *minio.Core
}

func NewClient(ossConf OssConf) (*OssClient, error) {
	opt := &minio.Options{
		Creds:  credentials.NewStaticV4(ossConf.AccessKeyID, ossConf.AecretAccessKey, ""),
		Secure: ossConf.Secure,
	}

	client, err := minio.New(ossConf.AliasedURL, opt)
	if err != nil {
		return nil, err
	}

	coreClient, err := minio.NewCore(ossConf.AliasedURL, opt)
	if err != nil {
		return nil, err
	}

	return &OssClient{
		conf:       oConf,
		client:     client,
		coreClient: coreClient,
	}, nil
}

func (oc *OssClient) GenMultiPartSignedUrl(ctx context.Context, oid string, chunkNum int64) (string, []string, error) {
	object := strings.TrimPrefix(path.Join(oc.conf.BasePath, path.Join(oid[0:2], oid[2:4], oid[4:])), "/")
	uploadId, err := oClient.coreClient.NewMultipartUpload(ctx, oc.conf.Bucket, object, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		return "", nil, err
	}

	reqParams := make(url.Values)
	reqParams.Set("uploadId", uploadId)

	// 碎片 ?
	urls := []string{}
	start := int64(1)
	for start <= chunkNum {
		reqParams.Set("partNumber", strconv.FormatInt(start, 10))
		url, err := oc.client.Presign(ctx, http.MethodPut, oc.conf.Bucket, object, PresignedUploadPartUrlExpireTime, reqParams)
		if err != nil {
			return "", nil, err
		}
		urls = append(urls, url.String())
		start += 1
	}
	return uploadId, urls, nil
}

func (oc *OssClient) CompleteMultiPartUpload(ctx context.Context, oid string, uploadID string, parts []Part) (string, error) {
	object := strings.TrimPrefix(path.Join(oc.conf.BasePath, path.Join(oid[0:2], oid[2:4], oid[4:])), "/")
	var mcp []minio.CompletePart
	for _, p := range parts {
		mcp = append(mcp, minio.CompletePart{ETag: p.ETag, PartNumber: p.PartNumber})
	}
	result, err := oc.coreClient.CompleteMultipartUpload(ctx, oc.conf.Bucket, object, uploadID, mcp, minio.PutObjectOptions{})
	return result.ETag, err
}
