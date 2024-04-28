package main

import (
	"fmt"
	"os"
	"path"

	"github.com/codecrafters-io/git-starter-go/object"
)

// Usage: your_git.sh <command> <arg1> <arg2> ...

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>....]\n")
	}

	switch command := os.Args[1]; command {
	case "init":
		var gitInitFiles = [3]string{".git", ".git/objects", ".git/refs"}

		for _, dir := range gitInitFiles {
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
			fmt.Fprintln(os.Stderr, "usage: mygit cat-file -p <object_id>")
			os.Exit(1)
		}

		if len(os.Args[3]) < 40 {
			fmt.Fprintf(os.Stderr, "Not a valid object name %s\n", os.Args[3])
			os.Exit(1)
		}

		content, err := catFile(os.Args[3])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		fmt.Print(content)

	case "hash-object":
		if len(os.Args) <= 3 || os.Args[2] != "-w" {
			fmt.Fprintln(os.Stderr, "usage: mygit hash-object -w <filepath>")
			os.Exit(1)
		}

		objectID, err := hashFile(os.Args[3])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		fmt.Println(objectID)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}

func catFile(objectID string) (string, error) {
	blobPath := path.Join(".git/objects", objectID[:2], objectID[2:])
	compressedFile, err := os.Open(blobPath)
	if err != nil {
		return "", fmt.Errorf("object %s not found", blobPath)
	}
	defer compressedFile.Close()

	blob, err := object.GetBlob(compressedFile)
	if err != nil {
		return "", err
	}

	return blob.Content, nil
}

func hashFile(filepath string) (string, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("object %s not found", filepath)
	}
	defer f.Close()

	return object.CreateBlob(f)
}
