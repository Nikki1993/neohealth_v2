package main

import (
	"crypto/md5"
	"embed"
	"encoding/hex"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	_ "net/http/pprof"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

//go:embed all:static/*
var Content embed.FS

//go:embed all:templates/*
var templates embed.FS

var tmpl = template.Must(parseTemplates())
var web = map[string]Website{"en": {}, "fi": {}, "ru": {}}
var hash = md5.New()
var md5sums = map[string]string{}

type neuteredFileSystem struct {
	fs http.FileSystem
}

func generateMD5sums(path string, sums *map[string]string) {
	dir, err := Content.ReadDir(path)
	if err != nil {
		log.Fatalf("Error reading directory to generate hashes %s", err)
	}

	for _, entry := range dir {
		name := entry.Name()
		if entry.IsDir() {
			generateMD5sums(path+"/"+name, sums)
			continue
		}

		file, err := Content.ReadFile(path + "/" + name)
		if err != nil {
			log.Fatalf("Error reading file to generate hash %v", err)
		}

		_, err = hash.Write(file)
		if err != nil {
			log.Fatalf("Error writing file hash to hashmap %v", err)
		}

		(*sums)[name] = hex.EncodeToString(hash.Sum(nil))
	}
}

func main() {
	dir, err := Content.ReadDir("static")
	if err != nil {
		panic(err)
	}

	generateMD5sums("static", &md5sums)
	css := "static/output.css"

	for _, entry := range dir {
		if !strings.Contains(entry.Name(), "style") {
			continue
		}

		css = "static/" + entry.Name()
		break
	}

	for k := range web {
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
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	cont, err := fs.Sub(fs.FS(Content), "static")
	if err != nil {
		panic(err)
	}

	static := http.StripPrefix(
		"/static/",
		Cache(http.FileServer(neuteredFileSystem{http.FS(cont)})),
	)

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

	s := len(files)
	filePaths := make([]string, s)
	for i, file := range files {
		filePaths[i] = "templates/" + file.Name()
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

		match := r.Header.Get("If-None-Match")
		if match == "" {
			h.ServeHTTP(w, r)
			return
		}

		if !strings.Contains(match, etag) {
			h.ServeHTTP(w, r)
			return
		}

		w.WriteHeader(http.StatusNotModified)
	}

	return http.HandlerFunc(fn)
}

func (n neuteredFileSystem) Open(p string) (http.File, error) {
	f, err := n.fs.Open(path.Clean(p))
	if err != nil {
		return nil, err
	}

	defer func(f http.File) {
		_ = f.Close()
	}(f)

	s, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if !s.IsDir() {
		return f, nil
	}

	if _, err := n.fs.Open(filepath.Join(p, "index.html")); err != nil {
		return nil, err
	}

	return f, nil
}
