package main

import (
	"fmt"
	"net/http"
	"strings"
)

func routes(w http.ResponseWriter, r *http.Request) {
	println(r.URL.Path)
	println(r.Method)
	if r.Method == http.MethodGet &&
		strings.HasPrefix(r.URL.Path, "blobs") &&
		strings.HasSuffix(r.URL.Path, "info") {
		BlobInfo(w, r)
	} else if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "blobs") {
		BlobNew(w, r)
	} else if r.Method == http.MethodPatch && strings.Contains(r.URL.Path, "blobs") {
		BlobUpload(w, r)
	} else if r.Method == http.MethodPut && strings.Contains(r.URL.Path, "blobs") {
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
