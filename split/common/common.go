package common

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const MAX_CHUNK_SIZE = int64(1024 * 1024 * 10)

const DATA_ROOT = "./data"
const META_DIR = "meta"
const CHUNK_DIR = "chunk"
const META_FILE = "meta"

type BlobNewRequest struct {
	Hash       string `json:"blob_hash"`
	HashMethod string `json:"hash_method"`
	TotalSize  int64  `json:"total_szie"`
}

type BlobNewResponse struct {
	UploadToken  string `json: "upload_token"`
	MaxChunkSize int64  `json:"max_chunk_size"`
}

type BlobPatchResponse struct {
	ChunkIndex string `json: "chunk_index"`
	ChunkSize  string `json: "chunk_size"`
	ChunkHash  string `json: "chunk_hash"`
	HashMethod string `json:"hash_method"`
}

func NewUploadToken(hashMethod string, blobHash string, totalSize int64) string {
	h := sha256.New()
	h.Write([]byte(hashMethod))
	h.Write([]byte(blobHash))
	h.Write([]byte(strconv.FormatInt(totalSize, 10)))
	h.Write([]byte(time.Now().UTC().Format("20060102150405")))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func TempSaveChunk(chunk io.Reader) (string, string, error) {
	hw := sha256.New()
	file, err := ioutil.TempFile(os.TempDir(), "chunk-*.dat")
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	w := io.MultiWriter(file, hw)
	_, err = io.Copy(w, chunk)
	if err != nil {
		return file.Name(), "", err
	}

	return file.Name(), fmt.Sprintf("%x", hw.Sum(nil)), nil
}

func init() {
	chunkDir := filepath.Join(DATA_ROOT, CHUNK_DIR)
	err := os.MkdirAll(chunkDir, os.ModePerm)
	if err != nil {
		panic(err)
	}
}

func touchFile(p string) error {
	_, err := os.Create(p)
	return err
}

func SaveChunk(tempChunk string, uploadToken, chunkHash string, chunkIndex string) error {
	targetChunk := filepath.Join(DATA_ROOT, CHUNK_DIR, chunkHash)
	err := os.Rename(tempChunk, targetChunk)
	if err != nil {
		return err
	}

	metaDir := filepath.Join(DATA_ROOT, META_DIR, uploadToken)
	icFile := filepath.Join(metaDir, fmt.Sprintf("ic-%s-%s", chunkIndex, chunkHash))
	touchFile(icFile)
	if err != nil {
		return err
	}

	ciFile := filepath.Join(metaDir, fmt.Sprintf("ci-%s-%s", chunkHash, chunkIndex))
	err = touchFile(ciFile)
	return err
}

func SaveMeta(uploadToken string, metaData io.Reader) error {
	// create init dir
	metaDir := filepath.Join(DATA_ROOT, META_DIR, uploadToken)
	err := os.MkdirAll(metaDir, os.ModePerm)
	if err != nil {
		return err
	}

	// create meta file
	metaFile := filepath.Join(metaDir, META_FILE)
	meta, err := os.Create(metaFile)
	if err != nil {
		return err
	}
	defer meta.Close()

	// save file
	_, err = io.Copy(meta, metaData)
	return err
}

func LoadMeta(uploadToken string) ([]byte, error) {
	metaDir := filepath.Join(DATA_ROOT, META_DIR, uploadToken, META_FILE)
	return ioutil.ReadFile(metaDir)
}

func TempRetrieveBlob(uploadToken string) (mergeFile string, mergeFileSha256 string, err error) {
	file, err := ioutil.TempFile(os.TempDir(), "merged-*.dat")
	if err != nil {
		return
	}
	defer func() {
		if file != nil {
			file.Close()
		}
		if err != nil {
			os.Remove(file.Name())
		}
	}()

	mfs, err := RetrieveBlob(uploadToken, file.Name())
	if err != nil {
		return
	}

	return file.Name(), mfs, err
}

func RetrieveBlob(uploadToken string, result string) (string, error) {
	metaDir := filepath.Join(DATA_ROOT, META_DIR, uploadToken)
	// find
	files := []string{}
	if err := filepath.Walk(metaDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasPrefix(info.Name(), "ic-") {
				files = append(files, path)
			}
			return nil
		}); err != nil {
		return "", err
	}

	// sort
	sort.Slice(files, func(i, j int) bool {
		return files[i] < files[j]
	})

	// merge
	mergeFileSha256, err := mergeV1(files, result)
	fmt.Println(mergeFileSha256)
	return mergeFileSha256, err
}

func Sha256(f io.Reader) string {
	h := sha256.New()
	io.Copy(h, f)
	return fmt.Sprintf("%x", h.Sum(nil))
}
