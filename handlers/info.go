package handlers

import (
	"net/http"
)

func HandleAbout(w http.ResponseWriter, r *http.Request) {
	data := GetPageData(r)
	RenderTemplate(w, "about", data)
}

func HandlePrivacy(w http.ResponseWriter, r *http.Request) {
	data := GetPageData(r)
	RenderTemplate(w, "privacy", data)
}

func HandleFaq(w http.ResponseWriter, r *http.Request) {
	data := GetPageData(r)
	RenderTemplate(w, "faq", data)
}

func HandleTerms(w http.ResponseWriter, r *http.Request) {
	data := GetPageData(r)
	RenderTemplate(w, "terms", data)
}
