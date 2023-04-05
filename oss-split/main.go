package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var oClient *OssClient
var oConf OssConf

func init() {
	oConf = OssConf{
		AliasedURL:      "127.0.0.1:9000",
		AccessKeyID:     "minioadmin",
		AecretAccessKey: "minioadmin",
		Bucket:          "test-1024",
		BasePath:        "mp",
		Location:        "us-east-1",
		Secure:          false,
	}
	oClient, _ = NewClient(oConf)
}

func writeResponse(w http.ResponseWriter, resp interface{}) {
	data, err := json.Marshal(resp)
	if err != nil {
		log.Fatalln(err)
	}
	w.Write(data)
}

func newMultipart(w http.ResponseWriter, r *http.Request) {
	var br BatchRequest
	err := json.NewDecoder(r.Body).Decode(&br)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println(br)

	bResp := BatchResponse{
		Transfer: "multipart",
		Objects:  []*ObjectResponse{},
	}

	for _, o := range br.Objects {
		// check
		if o.Size > MaxBlobSize {
			bResp.Objects = append(bResp.Objects, &ObjectResponse{
				Error: &ObjectError{
					Code:    http.StatusUnprocessableEntity,
					Message: "object size is to large to upload",
				},
			})
			continue
		}

		// create
		chunkNum := int64(math.Ceil(float64(o.Size) / ChunkSize))
		cn := int(o.Size % ChunkSize)
		if cn != 0 {
			chunkNum += 1
		}

		uploadID, urls, err := oClient.GenMultiPartSignedUrl(r.Context(), o.Oid, chunkNum)
		if err != nil {
			fmt.Println(err.Error())
			bResp.Objects = append(bResp.Objects, &ObjectResponse{
				Error: &ObjectError{
					Code:    http.StatusCreated,
					Message: "gen object multipart url failed",
				},
			})
			continue
		}

		link := &Link{
			Href:   "completed",
			Header: map[string]string{},
		}

		link.Header["chunk_size"] = strconv.FormatInt(ChunkSize, 10)
		link.Header["upload_id"] = uploadID
		for idx, url := range urls {
			link.Header[fmt.Sprintf("%04d", idx+1)] = url
		}

		bResp.Objects = append(bResp.Objects, &ObjectResponse{
			Pointer: Pointer{
				Oid:  o.Oid,
				Size: o.Size,
			},
			Actions: map[string]*Link{"upload": link},
			Error:   nil,
		})
	}

	writeResponse(w, bResp)
}

func completeMultipart(w http.ResponseWriter, r *http.Request) {
	var cmp CompleteMultipart
	err := json.NewDecoder(r.Body).Decode(&cmp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("%+v\n", cmp)

	etag, err := oClient.CompleteMultiPartUpload(r.Context(), cmp.Oid, cmp.UploadID, cmp.Parts)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(etag))
}

func main() {
	if os.Args[len(os.Args)-1] == "client" {
		RunClient()
		return
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	r.Route("/oss", func(r chi.Router) {
		r.Post("/new_multipart", newMultipart)
		r.Post("/complete_multipart", completeMultipart)
	})
	http.ListenAndServe(":9999", r)
}
