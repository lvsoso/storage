package common

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func mergeV1(files []string, merged string) (string, error) {
	realFiles := make([]string, len(files))
	for idx := range files {
		f := files[idx][len(files[idx])-64:]
		realFiles[idx] = filepath.Join(DATA_ROOT, CHUNK_DIR, f)
	}
	fmt.Println(realFiles)

	mergedFile, err := os.Create(merged)
	if err != nil {
		return "", err
	}
	defer mergedFile.Close()

	hasher := sha256.New()
	w := io.MultiWriter(mergedFile, hasher)

	for _, file := range realFiles {
		f, err := os.OpenFile(file, os.O_RDONLY, 0644)
		if err != nil {
			return "", err
		}
		// TODO: handle size
		io.Copy(w, f)
		err = f.Close()
		if err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}
