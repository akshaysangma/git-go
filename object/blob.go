package object

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"strconv"
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

	objectID, err := generateHash(data)
	if err != nil {
		return "", err
	}

	if err := createObjectFile(objectID, data); err != nil {
		return "", err
	}

	return objectID, nil
}

func GetBlob(reader io.Reader) (*Blob, error) {

	r, err := zlib.NewReader(reader)
	if err != nil {
		return &Blob{}, fmt.Errorf("object incorrect format: unable to decommpress")
	}
	defer r.Close()

	var decompressed bytes.Buffer

	io.Copy(&decompressed, r)

	headerBuf, err := decompressed.ReadBytes(spaceByte)
	if err != nil {
		return &Blob{}, fmt.Errorf("object incorrect format: unable to find space byte")
	}

	header := string(headerBuf)

	sizeBuf, err := decompressed.ReadBytes(nullByte)
	if err != nil {
		return &Blob{}, fmt.Errorf("object incorrect format: unable to find null byte")
	}

	size, err := strconv.ParseInt(string(bytes.Trim(sizeBuf, string(nullByte))), 10, 64)
	if err != nil {
		return &Blob{}, fmt.Errorf("object incorrect format: size is incorrect: %w", err)
	}

	contentBuffer := make([]byte, size)
	n, err := decompressed.Read(contentBuffer[:])
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

func (b *Blob) String() string {
	return fmt.Sprint(b.Content)
}
