package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
)

var staticPath = "/static/"

//go:embed static/*
var content embed.FS

//go:embed all:templates/*
var templates embed.FS

type neuteredFileSystem struct {
	fs http.FileSystem
}

func main() {
	cont, err := fs.Sub(fs.FS(content), "static")
	if err != nil {
		log.Fatalln(err)
	}

	static := http.StripPrefix(staticPath, http.FileServer(neuteredFileSystem{http.FS(cont)}))

	http.Handle(staticPath, static)
	http.HandleFunc("/", servePage)

	err = http.ListenAndServe(":1234", logRequest(http.DefaultServeMux))
	if err != nil {
		log.Fatalln("Error starting server on port :1234")
	}
}

func servePage(w http.ResponseWriter, r *http.Request) {
	web, err := GenerateTranslations(r)
	if err != nil {
		log.Fatalln(err)
	}

	tmpl, err := parseTemplates()
	if err != nil {
		log.Fatalln(err)
	}

	err = tmpl.ExecuteTemplate(w, "index.gohtml", web)
	if err != nil {
		log.Fatalln(err)
	}
}

func parseTemplates() (*template.Template, error) {
	var filePaths []string

	files, err := fs.ReadDir(templates, "templates")
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		filePaths = append(filePaths, fmt.Sprintf("templates/%v", file.Name()))
	}

	tmpl, err := template.ParseFS(templates, filePaths...)
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if s.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err := nfs.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}

			return nil, err
		}
	}

	return f, nil
}
