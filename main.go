package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

var staticPath = "/static/"

func main() {
	static := http.StripPrefix(staticPath, http.FileServer(http.Dir(fmt.Sprintf(".%s", staticPath))))

	http.Handle(staticPath, static)
	http.HandleFunc("/", servePage)

	err := http.ListenAndServe(":1234", logRequest(http.DefaultServeMux))
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

	files, err := os.ReadDir("./templates")
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		filePaths = append(filePaths, fmt.Sprintf("./templates/%v", file.Name()))
	}

	tmpl, err := template.ParseFiles(filePaths...)
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
