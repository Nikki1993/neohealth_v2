package main

import (
	"embed"
	"encoding/json"
	"golang.org/x/text/language"
	"html/template"
	"net/http"
	"os"
	"strings"
)

type Service struct {
	Description string `json:"description"`
	Name        string `json:"name"`
	Price       string `json:"price"`
}

type Card struct {
	Title       string    `json:"title"`
	SubTitle    string    `json:"subtitle"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Services    []Service `json:"prices"`
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
	Name    string   `json:"name"`
	Title   string   `json:"title"`
	Socials []Social `json:"socials"`
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

type Contact struct {
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
	Contacts  []Contact
}

var matcher = language.NewMatcher([]language.Tag{
	language.BritishEnglish,
	language.AmericanEnglish,
	language.English,
	language.Finnish,
	language.Russian,
})

const staticIconPath = "./static/icons/"
const staticPathLanguages = staticIconPath + "languages"
const staticPathBrands = staticIconPath + "brands"

//go:embed all:translations/*
var translations embed.FS

func getLang(r *http.Request) (string, string) {
	query := r.URL.Query().Get("lang")
	accept := r.Header.Get("Accept-Language")
	tag, _ := language.MatchStrings(matcher, query, accept)
	return tag.String(), getExtension(tag)
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
	imgs, err := os.ReadDir(staticPathLanguages)
	if err != nil {
		return nil, err
	}

	languages := make([]Language, 0, len(imgs))

	for _, img := range imgs {
		name := img.Name()
		imgName := strings.ReplaceAll(name, ".", "")
		languages = append(languages, Language{
			Href: "/?lang=" + imgName,
			Alt:  imgName,
			Src:  staticPathLanguages + "/" + name,
		})
	}

	return languages, nil
}

func generateBrandImgs() ([]string, error) {
	imgs, err := os.ReadDir(staticPathBrands)
	if err != nil {
		return nil, err
	}

	collection := make([]string, 0, len(imgs))

	for _, img := range imgs {
		collection = append(collection, img.Name())
	}

	return collection, nil
}

func GenerateTranslations(r *http.Request) (Website, error) {
	tag, ext := getLang(r)

	languages, err := generateLanguageLinks()
	if err != nil {
		return Website{}, err
	}

	intro := Intro{
		Top:    "A new beginning",
		Middle: "Neohealth",
		Bottom: "For your skin",
	}

	services, err := parseJSON[[3]Card]("services/" + ext)
	if err != nil {
		return Website{}, err
	}

	brands, err := parseJSON[Brands]("brands/" + ext)
	if err != nil {
		return Website{}, err
	}

	brands.Imgs, err = generateBrandImgs()
	if err != nil {
		return Website{}, err
	}

	about, err := parseJSON[About]("about/" + ext)
	if err != nil {
		return Website{}, err
	}

	about.Team, err = parseJSON[[]Member]("team/" + ext)
	if err != nil {
		return Website{}, err
	}

	contact, err := parseJSON[[]Contact]("contact/" + ext)
	if err != nil {
		return Website{}, err
	}

	return Website{
		Lang:      tag,
		Languages: languages,
		Intro:     intro,
		Services:  services,
		Brands:    brands,
		About:     about,
		Contacts:  contact,
	}, nil

}

func parseJSON[S any](p string) (S, error) {
	var res S

	bytes, err := translations.ReadFile("translations/" + p)
	if err != nil {
		return res, err
	}

	err = json.Unmarshal(bytes, &res)
	if err != nil {
		return res, err
	}

	return res, nil
}
