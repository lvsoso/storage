package storage

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const (
	MULTIPART_THRESHOLD_SIZE_BYTES = 50 * 1024 * 1024 // 50MB
	MULTIPART_SIZE_BYTES           = 50 * 1024 * 1024 // 50MB
)

var _ Storage = (*AWS)(nil)

type AWS struct {
	cfg AWSConfig
	svc *s3.Client
}

type AWSConfig struct {
	SecretAccessKey string
	AccessKey       string
	Endpoint        string
	Token           string
	Bucket          string
	Region          string
}

// newAWS  creates a storage backend of aws s3.
//
// Config:
//
//	{
//		"access_key_id": "NX0NYXIARMDS1AUN3E2I",
//		"secret_access_key": "QKB0KWZEINFXLYSQDCWZYJZHUEYPMPOGCPIGAPXV",
//		"region": "us-east-1",
//		"endpoint": "https://oss.aws.s3",
//		"internal_endpoint": "https://oss-internal.aws.s3",
//		"bucket": "abc"
//	  }
func newAWS(cfg interface{}) (*AWS, error) {
	config, ok := cfg.(map[string]string)
	if !ok {
		return nil, errInvalidConfig
	}

	awsCfg := AWSConfig{}
	awsCfg.AccessKey = config["access_key_id"]
	awsCfg.SecretAccessKey = config["secret_access_key"]
	awsCfg.Bucket = config["bucket"]
	awsCfg.Region = config["region"]
	awsCfg.Endpoint = config["endpoint"]

	if awsCfg.AccessKey == "" ||
		awsCfg.SecretAccessKey == "" ||
		awsCfg.Bucket == "" ||
		awsCfg.Region == "" ||
		awsCfg.Endpoint == "" {
		return nil, errInvalidConfig
	}

	creds := credentials.NewStaticCredentialsProvider(awsCfg.AccessKey, awsCfg.SecretAccessKey, "")
	svc := s3.NewFromConfig(aws.Config{
		BaseEndpoint: aws.String(awsCfg.Endpoint),
		Region:       *aws.String(awsCfg.Region),
		Credentials:  creds,
	})

	return &AWS{
		cfg: awsCfg,
		svc: svc,
	}, nil
}

func (as *AWS) Get(ctx context.Context, path string, opt ...Opt) (io.ReadCloser, error) {
	key := as.getKeyPath(path)
	params := &s3.GetObjectInput{
		Bucket: aws.String(as.cfg.Bucket),
		Key:    aws.String(key),
	}
	resp, err := as.svc.GetObject(ctx, params)
	if err != nil {
		defer resp.Body.Close()
		return nil, err
	}
	return resp.Body, nil
}

func (as *AWS) Put(ctx context.Context, path string, r io.ReadCloser, size int64, opt ...Opt) (int64, error) {
	if size > MULTIPART_THRESHOLD_SIZE_BYTES {
		return as.putMultipart(ctx, path, r, size, opt...)
	}

	key := as.getKeyPath(path)

	params := &s3.PutObjectInput{
		Bucket: aws.String(as.cfg.Bucket),
		Key:    aws.String(key),
		Body:   r,
	}

	_, err := as.svc.PutObject(ctx, params)
	if err != nil {
		return 0, err
	}
	return size, nil
}

func (as *AWS) putMultipart(ctx context.Context, path string, r io.ReadCloser, size int64, opt ...Opt) (int64, error) {
	key := as.getKeyPath(path)
	uploader := manager.NewUploader(as.svc, func(u *manager.Uploader) {
		u.PartSize = MULTIPART_SIZE_BYTES
	})
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(as.cfg.Bucket),
		Key:    aws.String(key),
		Body:   r,
	})
	if err != nil {
		return 0, err
	}
	return size, nil
}

func (as *AWS) Delete(ctx context.Context, path string, opt ...Opt) error {
	key := as.getKeyPath(path)
	params := &s3.DeleteObjectInput{
		Bucket: aws.String(as.cfg.Bucket),
		Key:    aws.String(key),
	}
	_, err := as.svc.DeleteObject(ctx, params)
	return err
}

func (as *AWS) Exist(ctx context.Context, path string) (bool, error) {
	key := as.getKeyPath(path)
	_, err := as.svc.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(as.cfg.Bucket),
		Key:    aws.String(key),
	})

	if err != nil {

		var noSuchKey *s3types.NoSuchKey
		if errors.As(err, &noSuchKey) {
			return false, nil
		}

		var noFound *s3types.NotFound
		if errors.As(err, &noFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (as *AWS) getKeyPath(path string) string {
	if strings.HasPrefix("/"+as.cfg.Bucket, path) {
		return path[len("/"+as.cfg.Bucket+"/")+1:]
	}
	return path[len(as.cfg.Bucket+"/")+1:]
}
