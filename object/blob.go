package object

import (
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
)

const (
	headerEnd = byte(' ')
	nullByte  = byte(0)
)

type Blob struct {
	Header  string
	Size    int64
	Content string
}

func CreateBlob(reader io.Reader) (string, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	data := []byte(fmt.Sprintf("blob %d%s%s", len(content), string(nullByte), string(content)))

	hasher := sha1.New()
	_, err = hasher.Write(data)
	if err != nil {
		return "", err
	}
	sha1Hash := hasher.Sum(nil)
	objectID := fmt.Sprintf("%x", sha1Hash)

	if err = os.Mkdir(path.Join(".git/objects", string(objectID[:2])), 0755); err != nil {
		return "", err
	}

	f, err := os.Create(path.Join(".git/objects", string(objectID[:2]), string(objectID[2:])))
	if err != nil {
		return "", err
	}
	defer f.Close()

	zlibWriter := zlib.NewWriter(f)
	defer zlibWriter.Close()
	_, err = zlibWriter.Write(data)
	if err != nil {
		return "", err
	}

	return string(objectID), nil
}

func GetBlob(reader io.Reader) (*Blob, error) {

	r, err := zlib.NewReader(reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Object incorrect format")
		os.Exit(1)
	}
	defer r.Close()

	var buf []byte
	buffer := make([]byte, 1)

	for {
		n, err := r.Read(buffer)
		if err != nil && err != io.EOF {
			return &Blob{}, fmt.Errorf("error reading object data: %w", err)
		}

		if n == 0 || buffer[0] == headerEnd {
			break
		}

		buf = append(buf, buffer...)
	}

	header := string(buf)
	buf = nil

	if header != "blob" {
		return &Blob{}, fmt.Errorf("invalid Object type : %s", header)
	}

	for {
		n, err := r.Read(buffer)
		if err != nil && err != io.EOF {
			return &Blob{}, fmt.Errorf("error reading object data: %w", err)
		}

		if n == 0 || buffer[0] == nullByte {
			break
		}

		buf = append(buf, buffer...)
	}

	s := string(buf)
	buf = nil

	size, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return &Blob{}, fmt.Errorf("error reading object data: %w", err)
	}

	contentBuffer := make([]byte, size)
	n, err := r.Read(contentBuffer)
	if err != nil && err != io.EOF {
		return &Blob{}, fmt.Errorf("error reading object data: %w", err)
	}

	if int64(n) < size {
		return &Blob{}, fmt.Errorf("object data corrupted")
	}

	content := string(contentBuffer)

	return &Blob{
		Header:  header,
		Size:    size,
		Content: content,
	}, nil
}
