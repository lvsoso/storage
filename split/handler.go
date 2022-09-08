package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type Blob struct {
	Hash        string
	HashMethod  string
	TotalSize   int64
	UUID        string
	ChunkSize   int64
	ChunkAmount int64
}

//BlobInit 分片上传初始化
// {"hash_method":"sha256", "blob_hash": "0x11121323", "total_szie": 5120, "chunk_size":512, "chunk_amount", 10}
func BlobInit(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	hashMethod := r.PostForm.Get("hash_method")
	blobHash := r.PostForm.Get("blob_hash")
	println(blobHash)
	totalSize, _ := strconv.ParseInt(r.PostForm.Get("total_size"), 10, 64)
	chunkSize, _ := strconv.ParseInt(r.PostForm.Get("chunk_size"), 10, 64)
	chunkAmount, _ := strconv.ParseInt(r.PostForm.Get("chunk_amount"), 10, 64)

	blob := &Blob{
		Hash:        blobHash,
		HashMethod:  hashMethod,
		TotalSize:   totalSize,
		ChunkSize:   chunkSize,
		ChunkAmount: chunkAmount,
		UUID:        "12345",
	}

	data, _ := json.Marshal(blob)
	w.Write(data)
}

// BlokUpload 分片上传
func BlobUpload(w http.ResponseWriter, r *http.Request) {

}

//BlobFinish 上传完成
func BlobFinish(w http.ResponseWriter, r *http.Request) {

}

// 分片情况
func BlobInfo(w http.ResponseWriter, r *http.Request) {

}
