package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/text/language"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type Item struct {
	Description string `json:"description"`
	Name        string `json:"name"`
	Price       string `json:"price"`
}

type Card struct {
	Title       string `json:"title"`
	SubTitle    string `json:"subtitle"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Services    []Item `json:"prices"`
}

type Language struct {
	Href string
	Src  string
	Alt  string
}

type Intro struct {
	Top    string
	Middle string
	Bottom string
}

type Brands struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Imgs        []string
}

type Member struct {
	Name  string
	Title string
}

type About struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Button      string `json:"button"`
	Extended    string `json:"extended"`
	Team        []Member
}

type Social struct {
	Icon string       `json:"icon"`
	Link template.URL `json:"link"`
	Name string       `json:"name"`
	Tag  string       `json:"tag"`
}

type Contacts struct {
	Title   string   `json:"title"`
	Socials []Social `json:"items"`
}

type Website struct {
	Lang      string
	Languages []Language
	Intro     Intro
	Services  [3]Card
	Brands    Brands
	About     About
	Contacts  Contacts
}

var matcher = language.NewMatcher([]language.Tag{
	language.BritishEnglish,
	language.AmericanEnglish,
	language.English,
	language.Finnish,
	language.Russian,
})

var staticIconPath = "./static/icons/"
var staticPathLanguages = staticIconPath + "languages"
var staticPathBrands = staticIconPath + "brands"

func getLang(r *http.Request) (language.Tag, int) {
	query := r.URL.Query().Get("lang")
	accept := r.Header.Get("Accept-Language")
	return language.MatchStrings(matcher, query, accept)
}

func getExtension(tag language.Tag) string {
	switch tag {
	case language.Finnish:
		return "fi.json"
	case language.Russian:
		return "ru.json"
	default:
		return "en.json"
	}
}

func generateLanguageLinks() ([]Language, error) {
	var languages []Language

	imgs, err := os.ReadDir(staticPathLanguages)
	if err != nil {
		return nil, err
	}

	for _, img := range imgs {
		imgName := strings.Split(img.Name(), ".")[0]
		languages = append(languages, Language{
			Href: fmt.Sprintf("/?lang=%s", imgName),
			Alt:  imgName,
			Src:  fmt.Sprintf("%s/%s", staticPathLanguages, img.Name()),
		})
	}

	return languages, nil
}

func generateBrandImgs() ([]string, error) {
	var collection []string

	imgs, err := os.ReadDir(staticPathBrands)
	if err != nil {
		return nil, err
	}

	for _, img := range imgs {
		collection = append(collection, img.Name())
	}

	return collection, nil
}

func GenerateTranslations(r *http.Request) (Website, error) {
	tag, _ := getLang(r)
	ext := getExtension(tag)

	languages, err := generateLanguageLinks()
	if err != nil {
		return Website{}, err
	}

	intro := Intro{
		Top:    "A new beginning",
		Middle: "Neohealth",
		Bottom: "For your skin",
	}

	services, err := parseJSON([3]Card{}, "services/"+ext)
	if err != nil {
		return Website{}, err
	}

	brands, err := parseJSON(Brands{}, "brands/"+ext)
	if err != nil {
		return Website{}, err
	}

	brands.Imgs, err = generateBrandImgs()
	if err != nil {
		return Website{}, err
	}

	about, err := parseJSON(About{}, "about/"+ext)
	if err != nil {
		return Website{}, err
	}

	about.Team = append(about.Team, Member{Name: "Elena Kononova", Title: "Founder / Doctor"})

	contact, err := parseJSON(Contacts{}, "contact/"+ext)
	if err != nil {
		return Website{}, err
	}

	return Website{
		Lang:      tag.String(),
		Languages: languages,
		Intro:     intro,
		Services:  services,
		Brands:    brands,
		About:     about,
		Contacts:  contact,
	}, nil

}

func parseJSON[S any](res S, p string) (S, error) {
	p = "./translations/" + p
	file, err := os.Open(p)
	if err != nil {
		return res, err
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(file)

	bytes, err := io.ReadAll(file)
	if err != nil {
		return res, err
	}

	err = json.Unmarshal(bytes, &res)
	if err != nil {
		return res, err
	}

	return res, nil
}
