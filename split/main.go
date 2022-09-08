package main

import (
	"fmt"
	"net/http"
	"strings"
)

func routes(w http.ResponseWriter, r *http.Request) {
	url := strings.Split(r.URL.Path, "?")[0]
	println(r.URL.Path)
	println(url)
	if r.Method == http.MethodGet &&
		strings.HasPrefix(url, "blobs") &&
		strings.HasSuffix(url, "info") {
		BlobInfo(w, r)
	} else if r.Method == http.MethodPost && strings.HasSuffix(url, "blobs") {
		BlobInit(w, r)
	} else if r.Method == http.MethodPatch && strings.HasSuffix(url, "blobs") {
		BlobUpload(w, r)
	} else if r.Method == http.MethodPut && strings.HasSuffix(url, "blobs") {
		BlobFinish(w, r)
	}
}

func main() {
	fmt.Println("start ...")
	http.HandleFunc("/", routes)
	err := http.ListenAndServe("0.0.0.0:9999", nil)
	if err != nil {
		fmt.Printf("Failed to start server, err:%s", err.Error())
	}
}
