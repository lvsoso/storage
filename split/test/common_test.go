package main

import (
	"os"
	"server/common"
	"testing"
)

func TestRetrieveBlob(t *testing.T) {
	os.Chdir("..")
	uploadToken := "a9f4c55ebf0182d31600f6aaac430aae20421340ef5d0b18f3c65545277c87ce"
	mergeFile, _, err := common.TempRetrieveBlob(uploadToken)
	if err != nil {
		t.Error(err)
	}
	os.Remove(mergeFile)
}
