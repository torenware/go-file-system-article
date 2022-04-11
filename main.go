package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

//go:embed files
var embeddedFiles embed.FS

func main() {
	filesDir := embeddedFiles

	// Make sure that directory listings
	// won't happen by making a directory
	// listable only via an index.html file:
	filteredDir := FilteringFS{
		fs: filesDir,
	}

	// Use a bit of middleware to filter out
	// dot files (see below)
	handler := wrappedFileServer(filteredDir)
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

// To block access to directory listings, wrap our file system
// with another filesystem that blocks them.

type FilteringFS struct {
	fs fs.FS
}

// And make the wrapper into an fs.FS by implementing its
// interface.
//
// This is updated from Alex Edward's article from 2018:
// @see https://www.alexedwards.net/blog/disable-http-fileserver-directory-listings
func (wrapper FilteringFS) Open(name string) (fs.File, error) {
	f, err := wrapper.fs.Open(name)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if s.IsDir() {
		// Have an index file or go home!
		index := filepath.Join(name, "index.html")
		if _, err := wrapper.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}

			return nil, err
		}
	}

	return f, nil
}
