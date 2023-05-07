// generate a minio presigned download function
// for a given bucket and object
//
// Usage:
// $ go run main.go <bucket> <object>
//
// Example:
// $ go run main.go my-bucket my-object
//
// Output:
//
//	func main() {
//	    // Initialize minio client object.
//	    minioClient, err := minio.New("play.min.io", "Q3AM3UQ867SPQQA43P2F",
//	        "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG")
//	    if err != nil {
//	        log.Fatalln(err)
//	    }
//	    // Generate presigned download URL for object my-object in my-bucket
//	    downloadURL, err := minioClient.PresignedGetObject("my-bucket", "my-object", time.Second*24*60*60, nil)
//	    if err != nil {
//	        log.Fatalln(err)
//	    }
//	    log.Println(downloadURL)
//	}
//
// NOTE: YOU MUST USE THE MINIO GO SDK TO RUN THIS PROGRAM.
//
//	$ go get -u github.com/minio/minio-go
//
// NOTE: This example assumes that you already have a running minio
// server at "play.min.io".  Update the values to match your setup.
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/minio/minio-go/v6"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalln("Usage: main <bucket> <object>")
	}
	bucket := os.Args[1]
	object := os.Args[2]

	// Initialize minio client object.
	minioClient, err := minio.New("127.0.0.1:9000", "minioadmin",
		"minioadmin", false)
	if err != nil {
		log.Fatalln(err)
	}

	// Generate presigned download URL for object my-object in my-bucket
	downloadURL, err := minioClient.PresignedGetObject(bucket, object, time.Second*24*60*60, nil)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(downloadURL)
}
