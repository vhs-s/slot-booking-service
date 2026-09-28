package handlers

import (
	"html/template"
	"log"
	"net/http"
)

func (sh *SystemHandler) IndexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFiles("web/template/header.html", "web/template/index.html", "web/template/footer.html")
		if err != nil {
			log.Println("Error parse template", err)
		}
		t.ExecuteTemplate(w, "index", nil)
	}
}
