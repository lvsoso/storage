package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type CompleteMultipart struct {
	UploadID string `json:"uploadId"`
	Oid      string `json:"oid"`
	Parts    []Part `json:"parts"`
}

type Part struct {
	PartNumber int
	ETag       string
}

func RunClient() {
	newMultipart := "http://127.0.0.1:9999/oss/new_multipart"
	completeMultipart := "http://127.0.0.1:9999/oss/complete_multipart"
	f, err := os.OpenFile("test/random_file", os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		log.Fatalln(err)
	}

	fsize := fi.Size()
	h := sha256.New()
	readData := make([]byte, 1024*1024)
	for {
		n, err := f.Read(readData)
		if err != nil {
			if err != io.EOF {
				log.Fatalln(err)
			}
			fmt.Println("EOF")
			break
		}
		_, err = h.Write(readData[:n])
		if err != nil {
			log.Fatalln(err)
		}
	}
	oid := fmt.Sprintf("%x", h.Sum(nil))
	fmt.Println(oid)
	brq := BatchRequest{
		Operation: "upload",
		Objects: []Pointer{
			{
				Oid:  oid,
				Size: fsize,
			},
		},
		Transfers: []string{"multipart", "basic"},
	}

	reqData, err := json.Marshal(brq)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := http.Post(newMultipart, "application/json", bytes.NewReader(reqData))
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	var brsp BatchResponse
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	err = json.Unmarshal(respData, &brsp)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("%+v\n", brsp)

	_, err = f.Seek(0, 0)
	if err != nil {
		log.Fatalln(err)
	}

	var parts []Part
	for _, o := range brsp.Objects {
		actionHeader := o.Actions["upload"].Header
		chunkSize, err := strconv.Atoi(actionHeader["chunk_size"])
		if err != nil {
			log.Fatalln(err)
		}
		delete(actionHeader, "chunk_size")
		uploadID := actionHeader["upload_id"]
		delete(actionHeader, "upload_id")

		urls := sortUrls(actionHeader)
		start := 0
		readData := make([]byte, chunkSize)
		for idx, url := range urls {
			seekFrom := start * chunkSize
			f.Seek(int64(seekFrom), 0)
			n, err := f.Read(readData)
			if err != nil {
				log.Fatalln(err)
			}

			etag, err := uploadData(url, bytes.NewReader(readData[:n]))
			if err != nil {
				log.Fatalln(err)
			}

			parts = append(parts, Part{
				ETag:       strings.Replace(etag, "\"", "", -1),
				PartNumber: idx + 1,
			})
			start += 1
		}

		fmt.Printf("%+v\n", parts)
		cmp := CompleteMultipart{
			UploadID: uploadID,
			Oid:      oid,
			Parts:    parts,
		}

		reqData, err := json.Marshal(cmp)
		if err != nil {
			log.Fatalln(err)
		}

		resp, err := http.Post(completeMultipart, "application/json", bytes.NewReader(reqData))
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(resp.StatusCode)
	}

}

func sortUrls(header map[string]string) []string {
	var urls []string
	size := len(header)
	start := 1
	for start <= size {
		k := fmt.Sprintf("%04d", start)
		urls = append(urls, header[k])
		start += 1
	}
	return urls
}

func uploadData(url string, data io.Reader) (string, error) {

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, url, data)
	req.Header.Set("Content-Type", "application/octet-stream")
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	return resp.Header.Get("Etag"), nil
}
