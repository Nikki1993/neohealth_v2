package main

import (
	"embed"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	_ "net/http/pprof"
	"path"
	"path/filepath"
	"time"
)

//go:embed all:static/*
var Content embed.FS

//go:embed all:templates/*
var templates embed.FS

var tmpl = template.Must(parseTemplates())
var web = map[string]Website{
	"en": {},
	"fi": {},
	"ru": {},
}

type neuteredFileSystem struct {
	fs http.FileSystem
}

func main() {
	for k, _ := range web {
		t, err := GenerateTranslations(k)
		if err != nil {
			panic(err)
		}

		web[k] = t
	}

	r := chi.NewRouter()
	r.Use(middleware.Compress(5))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	cont, err := fs.Sub(fs.FS(Content), "static")
	if err != nil {
		log.Fatalln(err)
	}

	static := http.StripPrefix("/static/", http.FileServer(neuteredFileSystem{http.FS(cont)}))

	r.Handle("/static/*", static)
	r.HandleFunc("/", serveWebpage)

	err = http.ListenAndServe(":1234", r)
	if err != nil {
		log.Fatalln("Error starting server on port :1234")
	}
}

func serveWebpage(w http.ResponseWriter, r *http.Request) {
	err := tmpl.ExecuteTemplate(w, "index.gohtml", web[getLang(r)])
	if err != nil {
		log.Fatalln(err)
	}
}

func parseTemplates() (*template.Template, error) {
	files, err := fs.ReadDir(templates, "templates")
	if err != nil {
		return nil, err
	}

	filePaths := make([]string, 0, len(files))

	for _, file := range files {
		filePaths = append(filePaths, "templates/"+file.Name())
	}

	tmpl, err := template.ParseFS(templates, filePaths...)
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

func (n neuteredFileSystem) Open(p string) (http.File, error) {
	f, err := n.fs.Open(path.Clean(p))
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if s.IsDir() {
		if _, err := n.fs.Open(filepath.Join(p, "index.html")); err != nil {
			fileError := f.Close()
			if fileError != nil {
				return nil, fileError
			}

			return nil, err
		}
	}

	return f, nil
}
