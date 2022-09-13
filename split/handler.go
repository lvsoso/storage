package main

import (
	"encoding/json"
	"net/http"
	"os"
	"server/common"
	"strings"
)

type Blob struct {
	Hash       string
	HashMethod string
	TotalSize  int64
}

func (b Blob) String() string {
	data, _ := json.Marshal(b)
	return string(data)
}

//BlobNew 分片上传初始化
// {"hash_method":"sha256", "blob_hash": "0x11121323", "total_szie": 5120}
func BlobNew(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var bnr common.BlobNewRequest
	err := decoder.Decode(&bnr)
	if err != nil {
		panic(err)
	}

	uploadToken := common.NewUploadToken(bnr.HashMethod, bnr.Hash, bnr.TotalSize)

	blob := &Blob{
		Hash:       bnr.Hash,
		HashMethod: bnr.HashMethod,
		TotalSize:  bnr.TotalSize,
	}
	err = common.SaveMeta(uploadToken, strings.NewReader(blob.String()))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := &common.BlobNewResponse{
		UploadToken:  uploadToken,
		MaxChunkSize: common.MAX_CHUNK_SIZE,
	}

	data, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
}

//BlobUpload ...
func BlobUpload(w http.ResponseWriter, r *http.Request) {
	t := strings.Split(r.URL.Path, "/")
	uploadToken := t[len(t)-1]

	err := r.ParseMultipartForm(common.MAX_CHUNK_SIZE + 1024*1024 + 1)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	chunkIndex := r.PostFormValue("chunk_index")
	chunkSize := r.PostFormValue("chunk_size")
	chunkHash := r.PostFormValue("chunk_hash")
	hashMethod := r.PostFormValue("hash_method")

	chunk, _, err := r.FormFile("chunk")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer chunk.Close()

	// save chunk temp
	tempFile, finnalHash, err := common.TempSaveChunk(chunk)
	defer func() {
		if f, err := os.OpenFile(tempFile, os.O_RDONLY, os.ModePerm); err == nil {
			if fi, err := f.Stat(); err == nil {
				os.Remove(fi.Name())
			}
		}

	}()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// check chunk hash
	if finnalHash != chunkHash {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// save chunk
	if err := common.SaveChunk(tempFile, uploadToken, chunkHash, chunkIndex); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := &common.BlobPatchResponse{
		ChunkIndex: chunkIndex,
		ChunkSize:  chunkSize,
		ChunkHash:  chunkHash,
		HashMethod: hashMethod,
	}
	data, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
}

//BlobFinish ...
func BlobFinish(w http.ResponseWriter, r *http.Request) {
	t := strings.Split(r.URL.Path, "/")
	uploadToken := t[len(t)-1]

	// merge files
	mergeFile, mfs, err := common.TempRetrieveBlob(uploadToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() {
		os.Remove(mergeFile)
	}()

	// get metadata
	metaData, err := common.LoadMeta(uploadToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	blob := &Blob{}
	if err := json.Unmarshal(metaData, blob); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// check hash
	if mfs != blob.Hash {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// 分片情况
func BlobInfo(w http.ResponseWriter, r *http.Request) {

}
