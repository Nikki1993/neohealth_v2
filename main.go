package main

import (
	"crypto/md5"
	"embed"
	"encoding/hex"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	_ "net/http/pprof"
	"path"
	"path/filepath"
	"strings"
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
var hash = md5.New()
var md5sums = map[string]string{}

type neuteredFileSystem struct {
	fs http.FileSystem
}

func generateMD5sums(path string, sums *map[string]string) {
	dir, err := Content.ReadDir(path)
	if err != nil {
		log.Println(err)
		return
	}

	for _, entry := range dir {
		name := entry.Name()
		if entry.IsDir() {
			generateMD5sums(path+"/"+name, sums)
			continue
		}

		file, err := Content.ReadFile(path + "/" + name)
		if err != nil {
			log.Println(err)
			continue
		}

		hash.Write(file)
		(*sums)[name] = hex.EncodeToString(hash.Sum(nil))
	}
}

func main() {
	dir, err := Content.ReadDir("static")
	if err != nil {
		log.Fatalln(err)
	}

	generateMD5sums("static", &md5sums)

	css := "static/output.css"

	for _, entry := range dir {
		if strings.Contains(entry.Name(), "style") {
			css = "static/" + entry.Name()
		}
	}

	for k, _ := range web {
		t, err := GenerateTranslations(k)
		if err != nil {
			panic(err)
		}

		t.Css = template.URL(css)
		web[k] = t
	}

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*"},
		AllowedMethods:   []string{"GET"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(middleware.Compress(5))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	//r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	cont, err := fs.Sub(fs.FS(Content), "static")
	if err != nil {
		log.Fatalln(err)
	}

	static := http.StripPrefix("/static/", Cache(http.FileServer(neuteredFileSystem{http.FS(cont)})))

	r.Handle("/static/*", static)
	r.HandleFunc("/", serveWebpage)

	err = http.ListenAndServe(":1234", r)
	if err != nil {
		log.Fatalln("Error starting server on port :1234")
	}
}

func serveWebpage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")
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

func Cache(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		s := strings.Split(r.URL.String(), "/")
		etag := md5sums[s[len(s)-1]]
		w.Header().Set("Cache-Control", "max-age=43200")
		w.Header().Set("ETag", etag)

		if match := r.Header.Get("If-None-Match"); match != "" {
			if strings.Contains(match, etag) {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}

		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
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
