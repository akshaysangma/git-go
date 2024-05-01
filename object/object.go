package object

import (
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"os"
	"path"
)

const (
	spaceByte  = byte(' ')
	nullByte   = byte(0)
	excludeDir = ".git"
)

func generateHash(data []byte) (string, error) {
	hasher := sha1.New()
	_, err := hasher.Write(data)
	if err != nil {
		return "", err
	}
	sha1Hash := hasher.Sum(nil)
	return fmt.Sprintf("%x", sha1Hash), nil
}

func createObjectFile(objectID string, data []byte) error {
	if err := os.Mkdir(path.Join(".git/objects", objectID[:2]), 0755); err != nil {
		return err
	}

	f, err := os.Create(path.Join(".git/objects", objectID[:2], objectID[2:]))
	if err != nil {
		return err
	}
	defer f.Close()

	zlibWriter := zlib.NewWriter(f)
	defer zlibWriter.Close()
	_, err = zlibWriter.Write(data)
	if err != nil {
		return err
	}

	return nil
}
