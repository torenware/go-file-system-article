package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
)

//go:embed files
var embeddedFiles embed.FS

func main() {
	// Here's a "safe" embed example
	filesDir := embeddedFiles

	// BUT... note we need to adjust where the "top"
	// of the file system is, as so:
	adjustedDir, err := fs.Sub(filesDir, "files")
	if err != nil {
		log.Fatal(err)
	}

	listFiles("", adjustedDir, ".")

	handler := http.FileServer(http.FS(adjustedDir))
	http.Handle("/", handler)

	log.Println("Serving static files at :5000")
	err = http.ListenAndServe(":5000", handler)
	if err != nil {
		log.Fatal(err)
	}
}

func listFiles(indent string, dir fs.FS, path string) error {
	items, err := fs.ReadDir(dir, path)
	if err != nil {
		log.Printf("could not list files for dir %s: %s", ".", err)
		return err
	}

	for _, item := range items {
		name := item.Name()
		if item.IsDir() {
			fmt.Println(indent, name+"/")
			subDir, err := fs.Sub(dir, name)
			if err != nil {
				return err
			}
			listFiles(indent+"    ", subDir, ".")
		} else {
			fmt.Println(indent, name)
		}
	}
	return nil
}
