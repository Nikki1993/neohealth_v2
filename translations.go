package main

import (
	"embed"
	"encoding/json"
	"golang.org/x/text/language"
	"html/template"
	"net/http"
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
	Team        [1]Member
}

type Social struct {
	Icon string       `json:"icon"`
	Link template.URL `json:"link"`
	Name string       `json:"name"`
	Tag  string       `json:"tag"`
}

type Contact struct {
	Title   string    `json:"title"`
	Socials [6]Social `json:"items"`
}

type Website struct {
	Lang      string
	Languages []Language
	Intro     Intro
	Services  [3]Card
	Brands    Brands
	About     About
	Contacts  Contact
}

var matcher = language.NewMatcher([]language.Tag{
	language.BritishEnglish,
	language.AmericanEnglish,
	language.English,
	language.Finnish,
	language.Russian,
})

const staticIconPath = "static/icons/"
const staticPathLanguages = staticIconPath + "languages"
const staticPathBrands = staticIconPath + "brands"

//go:embed all:translations/*
var translations embed.FS

func getLang(r *http.Request) string {
	query := r.URL.Query().Get("lang")
	accept := r.Header.Get("Accept-Language")
	tag, _ := language.MatchStrings(matcher, query, accept)

	switch tag {
	case language.Finnish:
		return "fi"
	case language.Russian:
		return "ru"
	default:
		return "en"
	}
}

func generateLanguageLinks() ([]Language, error) {
	imgs, err := Content.ReadDir(staticPathLanguages)
	if err != nil {
		return nil, err
	}

	languages := make([]Language, 0, len(imgs))

	for _, img := range imgs {
		name := img.Name()
		imgName := strings.ReplaceAll(name, ".svg", "")
		languages = append(languages, Language{
			Href: "/?lang=" + imgName,
			Alt:  imgName,
			Src:  staticPathLanguages + "/" + name,
		})
	}

	return languages, nil
}

func generateBrandImgs() ([]string, error) {
	imgs, err := Content.ReadDir(staticPathBrands)
	if err != nil {
		return nil, err
	}

	collection := make([]string, 0, len(imgs))

	for _, img := range imgs {
		collection = append(collection, img.Name())
	}

	return collection, nil
}

func GenerateTranslations(lang string) (Website, error) {
	website := Website{}
	website.Lang = lang

	languages, err := generateLanguageLinks()
	if err != nil {
		return website, err
	}

	website.Languages = languages

	website.Intro = Intro{
		Top:    "A new beginning",
		Middle: "Neohealth",
		Bottom: "For your skin",
	}

	ext := lang + ".json"

	website.Services, err = parseJSON[[3]Card]("services/" + ext)
	if err != nil {
		return website, err
	}

	website.Brands, err = parseJSON[Brands]("brands/" + ext)
	if err != nil {
		return website, err
	}

	website.Brands.Imgs, err = generateBrandImgs()
	if err != nil {
		return website, err
	}

	website.Contacts, err = parseJSON[Contact]("contact/" + ext)
	if err != nil {
		return website, err
	}

	return website, nil
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
