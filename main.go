package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
)

//go:embed files
var embeddedFiles embed.FS

func main() {
	filesDir := os.DirFS("files")

	listFiles("", filesDir, ".")

	// Use a bit of middleware to filter out
	// dot files (see below)
	handler := wrappedFileServer(filesDir)
	http.Handle("/", handler)

	log.Println("Serving static files at :5000")
	err := http.ListenAndServe(":5000", handler)
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

// Wrap file server and block dot files from appearing
func wrappedFileServer(root fs.FS) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Path
		// strip off the initial / if it's there
		if len(url) > 0 && url[:1] == "/" {
			url = url[1:]
		}
		path := strings.Split(url, "/")

		for _, stem := range path {
			// If it's a dot file, make it unseen
			if len(stem) > 0 && stem[:1] == "." {
				http.NotFound(w, r)
				return
			}
		}
		// We're using fs.FS and not http.FileSystem, so convert
		// with http.FS:
		fileServer := http.StripPrefix("/", http.FileServer(http.FS(root)))
		// and dispatch our approved files to that handler
		fileServer.ServeHTTP(w, r)
	}

	return http.HandlerFunc(handler)
}
