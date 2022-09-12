package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"server/common"
	"strconv"
	"strings"
	"time"
)

const jsonCT = "application/json"
const chunkSize = 1024 * 1024 * 4 // 4 MB

func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
}

func multipartUpload(filename string, targetURL string, chunkSize int) error {

	f, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer f.Close()

	bfRd := bufio.NewReader(f)
	index := 0

	client := &http.Client{}
	ch := make(chan int)
	buf := make([]byte, chunkSize)
	for {
		n, err := bfRd.Read(buf)
		if n <= 0 {
			break
		}
		index++

		bufCopied := make([]byte, chunkSize)
		copy(bufCopied, buf)
		h := sha256.New()
		h.Write(bufCopied[:n])
		bufHash := fmt.Sprintf("%x", h.Sum(nil))

		go func(b []byte, curIdx int) {
			fmt.Printf("upload_size: %d\n", len(b))

			// build multipart form
			values := map[string]io.Reader{
				"chunk":       bytes.NewReader(b),
				"chunk_index": strings.NewReader(strconv.Itoa(curIdx)),
				"chunk_size":  strings.NewReader(strconv.Itoa(len(b))),
				"chunk_hash":  strings.NewReader(bufHash),
				"hash_method": strings.NewReader("sha256"),
			}
			var buf bytes.Buffer
			w := multipart.NewWriter(&buf)
			for key, r := range values {
				var fw io.Writer
				if x, ok := r.(io.Closer); ok {
					defer x.Close()
				}
				if key == "chunk" {
					fw, err = w.CreateFormFile(key, fmt.Sprintf("%s.%d", filename, curIdx))
				} else {
					fw, err = w.CreateFormField(key)
				}
				if err != nil {
					return
				}
				if _, err = io.Copy(fw, r); err != nil {
					return
				}
			}
			w.Close()

			// do request
			req, err := http.NewRequest(http.MethodPatch, targetURL, &buf)
			if err != nil {
				return
			}
			req.Header.Set("Content-Type", w.FormDataContentType())
			resp, err := client.Do(req)
			if err != nil {
				return
			}

			// Check the response
			if resp.StatusCode != http.StatusOK {
				err = fmt.Errorf("bad status: %s", resp.Status)
				return
			}

			defer resp.Body.Close()
			decoder := json.NewDecoder(resp.Body)
			var bpresp common.BlobPatchResponse
			err = decoder.Decode(&bpresp)
			checkError(err)
			fmt.Println(bpresp)
			ch <- curIdx
		}(bufCopied[:n], index)

		if err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Println(err.Error())
			}
		}
	}

	for idx := 0; idx < index; idx++ {
		select {
		case res := <-ch:
			fmt.Println(res)
		default:
			time.Sleep(1 * time.Second)
		}
	}
	return nil
}

func main() {
	// init upload
	// file := "/home/lv/Downloads/archlinux-2020.06.01-x86_64.iso"
	file := "/tmp/3M.file"
	fileHash := "3ee1a72ceb192638ed08d9df27f525f3239c4f7c6e77c41c504669a87ee3caca"
	f, err := os.OpenFile(file, os.O_RDONLY, 0755)
	checkError(err)
	defer f.Close()

	fi, err := f.Stat()
	checkError(err)

	data, _ := json.Marshal(&common.BlobNewRequest{
		Hash:       fileHash,
		TotalSize:  fi.Size(),
		HashMethod: "sha256",
	})

	resp, err := http.Post("http://127.0.0.1:9999/v1/blobs", jsonCT, bytes.NewBuffer(data))
	checkError(err)
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var bnresp common.BlobNewResponse
	err = decoder.Decode(&bnresp)
	checkError(err)

	// split upload
	targetURL := fmt.Sprintf("http://127.0.0.1:9999/v1/blobs/%s", bnresp.UploadToken)
	err = multipartUpload(file, targetURL, chunkSize)
	checkError(err)

	// commit upload
	req, err := http.NewRequest(http.MethodPut, targetURL, strings.NewReader("{}"))
	checkError(err)
	client := http.Client{}
	resp, err = client.Do(req)
	checkError(err)
	if resp.StatusCode != http.StatusOK {
		panic(fmt.Errorf("bad status: %s", resp.Status))
	}

	println("finished........................................")
}
