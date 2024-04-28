package main

import (
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
)

// Usage: your_git.sh <command> <arg1> <arg2> ...

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>....]\n")
	}

	switch command := os.Args[1]; command {
	case "init":
		for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory %s : %s ", dir, err)
			}
		}
		headFileContents := []byte("ref: refs/heads/main\n")
		if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file %s : %s", ".git/HEAD", err)
		}

		fmt.Println("Initialized git directory")

	case "cat-file":
		if len(os.Args) <= 3 || os.Args[2] != "-p" {
			fmt.Fprintln(os.Stderr, "usage: mygit cat-file -p <blob_sha>")
			os.Exit(1)
		}

		if len(os.Args[3]) < 40 {
			fmt.Fprintf(os.Stderr, "Not a valid object name %s\n", os.Args[3])
			os.Exit(1)
		}

		blobPath := path.Join(".git/objects", os.Args[3][:2], os.Args[3][2:])
		compressedFile, err := os.Open(blobPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Object %s not found\n", blobPath)
			os.Exit(1)
		}
		defer compressedFile.Close()

		r, err := zlib.NewReader(compressedFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Object incorrect format")
			os.Exit(1)
		}
		defer r.Close()

		var buf []byte
		headerEnd := byte(' ')
		nullByte := byte(0)

		buffer := make([]byte, 1)
		for {
			n, err := r.Read(buffer)
			if err != nil && err != io.EOF {
				fmt.Fprintf(os.Stderr, "Error reading object data: %s\n", err)
				os.Exit(1)
			}

			if n == 0 || buffer[0] == headerEnd {
				break
			}

			buf = append(buf, buffer...)
		}

		header := string(buf)
		buf = nil

		if header != "blob" {
			fmt.Fprintf(os.Stderr, "Invalid Object type : %s\n", header)
			os.Exit(1)
		}

		for {
			n, err := r.Read(buffer)
			if err != nil && err != io.EOF {
				fmt.Fprintf(os.Stderr, "Error reading object data: %s\n", err)
				os.Exit(1)
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
			fmt.Fprintf(os.Stderr, "Error reading object data: %s\n", err)
			os.Exit(1)
		}

		contentBuffer := make([]byte, size)
		n, err := r.Read(contentBuffer)
		if err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "Error reading object data: %s\n", err)
			os.Exit(1)
		}

		if int64(n) < size {
			fmt.Fprintf(os.Stderr, "Object data corrupted\n")
			os.Exit(1)
		}

		content := string(contentBuffer)
		fmt.Fprint(os.Stdout, content)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
