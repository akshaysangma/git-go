package object

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
)

type TreeEntry struct {
	Mode string
	Name string
	Hash string
}

type Tree struct {
	Header  string
	Size    int64
	Content []*TreeEntry
}

func CreateTree(dir string) (string, error) {
	dirItems, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	content := ""

	for _, item := range dirItems {
		filepath := path.Join(dir, item.Name())
		if item.IsDir() {
			if item.Name() == excludeDir {
				continue
			}
			content += "40000 " + item.Name() + string(nullByte)

			hash, err := CreateTree(filepath)
			if err != nil {
				return "", err
			}

			hash20sha, err := hex.DecodeString(hash)
			if err != nil {
				return "", err
			}

			content += string(hash20sha)
		} else {
			content += "100644 " + item.Name() + string(nullByte)
			f, err := os.Open(filepath)
			if err != nil {
				return "", fmt.Errorf("object %s not found", filepath)
			}
			defer f.Close()
			hash, err := CreateBlob(f)
			if err != nil {
				return "", err
			}
			hash20sha, err := hex.DecodeString(hash)
			if err != nil {
				return "", err
			}

			content += string(hash20sha)
		}
	}
	output := "tree " + strconv.Itoa(len(content)) + string(nullByte) + content
	hash, err := generateHash([]byte(output))
	if err != nil {
		return "", err
	}

	err = createObjectFile(hash, []byte(output))
	if err != nil {
		return "", err
	}

	return hash, nil
}

func GetTree(reader io.Reader) (*Tree, error) {

	r, err := zlib.NewReader(reader)
	if err != nil {
		return &Tree{}, fmt.Errorf("object incorrect format: unable to decommpress")
	}
	defer r.Close()

	var decompressed bytes.Buffer

	io.Copy(&decompressed, r)

	headerBuf, err := decompressed.ReadBytes(spaceByte)
	if err != nil {
		return &Tree{}, fmt.Errorf("object incorrect format: unable to find space byte")
	}

	header := string(headerBuf)

	sizeBuf, err := decompressed.ReadBytes(nullByte)
	if err != nil {
		return &Tree{}, fmt.Errorf("object incorrect format: unable to find null byte")
	}

	size, err := strconv.ParseInt(string(bytes.Trim(sizeBuf, string(nullByte))), 10, 64)
	if err != nil {
		return &Tree{}, fmt.Errorf("object incorrect format: size is incorrect: %w", err)
	}

	var entries []*TreeEntry
	for buffer, err := decompressed.ReadBytes(nullByte); err == nil; buffer, err = decompressed.ReadBytes(nullByte) {
		buffer = bytes.Trim(buffer, string(nullByte))

		var entry TreeEntry
		buf := bytes.Split(buffer, []byte(" "))

		if len(buf) < 2 {
			return &Tree{}, fmt.Errorf("object incorrect format : less than 2")
		}

		entry.Mode = string(buf[0])
		entry.Name = string(buf[1])

		var hashBuffer [20]byte
		_, rErr := decompressed.Read(hashBuffer[:])
		if rErr == io.EOF {
			break
		}
		if rErr != nil {
			return &Tree{}, fmt.Errorf("object incorrect format : %w", rErr)
		}

		entry.Hash = string(hashBuffer[:])
		entries = append(entries, &entry)
	}

	return &Tree{
		Header:  header,
		Size:    size,
		Content: entries,
	}, nil
}

func (t *Tree) String() string {
	var sb strings.Builder
	for _, entry := range t.Content {
		if entry.Mode == "040000" {
			sb.WriteString(fmt.Sprintf("%s tree %x %s\n", entry.Mode, entry.Hash, entry.Name))
			continue
		}
		sb.WriteString(fmt.Sprintf("%s blob %x %s\n", entry.Mode, entry.Hash, entry.Name))
	}
	return sb.String()
}

func (t *Tree) NameOnlyString() string {
	var sb strings.Builder
	for _, entry := range t.Content {
		sb.WriteString(fmt.Sprintf("%s\n", entry.Name))
	}
	return sb.String()
}
