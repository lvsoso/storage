// https://doc.bscstorage.com/doc/s2/demo/go.html

package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var access_key = ""
var secret_key = ""
var token = ""
var end_point = ""
var bucket = ""
var testKey = ""

func main() {
	credential := credentials.NewStaticCredentials(access_key, secret_key, token)

	config := aws.NewConfig().WithRegion("us-east-1").
		WithEndpoint(end_point).
		WithCredentials(credential).WithS3ForcePathStyle(false)

	sess := session.New(config)
	svc := s3.New(sess)

	// uploader := s3manager.NewUploader(sess)
	// downloader := s3manager.NewDownloader(sess)

	// PUT obj
	// params := &s3.PutObjectInput{
	// 	Bucket: aws.String(bucket),
	// 	Key:    aws.String("test-key"),
	// 	// ACL:         aws.String("public-read"),
	// 	// ContentType: aws.String("image/jpeg"), //请替换为合适的文件类型
	// 	Body: bytes.NewReader([]byte("bla bla")),
	// 	// Metadata: map[string]*string{
	// 	// 	"key-foo": aws.String("value-bar"),
	// 	// },
	// }

	// resp, err := svc.PutObject(params)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }
	// fmt.Println(resp)

	// Upload file
	//
	// f, err := os.Open("test.txt")
	// if err != nil {
	// 	fmt.Println("open file error")
	// 	return
	// }

	// params := &s3manager.UploadInput{
	// 	Bucket: aws.String(bucket),
	// 	Key:    aws.String(testKey),
	// 	Body:   f,
	// }

	// result, err := uploader.Upload(params)
	// if err != nil {
	// 	fmt.Println("upload file error")
	// 	return
	// }
	// fmt.Printf("file uploaded to: %s\n", result.Location)

	// Get obj
	// params := &s3.GetObjectInput{
	// 	Bucket: aws.String(bucket),
	// 	Key:    aws.String(testKey),
	// }

	// resp, err := svc.GetObject(params)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }
	// buf := new(bytes.Buffer)
	// buf.ReadFrom(resp.Body)
	// fmt.Println(buf.String())

	// Download obj
	//
	// f, err := os.Create("test.txt.download")
	// if err != nil {
	// 	fmt.Println("create file error")
	// 	return
	// }

	// params := &s3.GetObjectInput{
	// 	Bucket: aws.String(bucket),
	// 	Key:    aws.String(testKey),
	// }

	// n, err := downloader.Download(f, params)
	// if err != nil {
	// 	fmt.Println("download file error")
	// 	return
	// }
	// fmt.Printf("file download %d bytes\n", n)

	//

	// get presign url
	// params := &s3.GetObjectInput{
	// 	Bucket: aws.String(bucket),
	// 	Key:    aws.String(testKey),
	// }
	// req, _ := svc.GetObjectRequest(params)
	// url, _ := req.Presign(300 * time.Second)
	// fmt.Println(url)

	marker := ""

	for {
		params := &s3.ListObjectsInput{
			Bucket: aws.String(bucket),
			Marker: aws.String(marker),
		}

		resp, err := svc.ListObjects(params)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		if len(resp.Contents) == 0 {
			break
		}

		for _, content := range resp.Contents {
			fmt.Printf("key:%s, size:%d\n", *content.Key, *content.Size)
			marker = *content.Key
		}
	}
}
