package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
)

type handler struct {
	r    repository
	tmpl *template.Template
}

func initHandlers(r repository) error {
	tmpl, err := template.ParseGlob("templates/*")
	if err != nil {
		return fmt.Errorf("parse templates: %w", err)
	}

	h := handler{r: r, tmpl: tmpl}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", h.handleIndex)
	mux.HandleFunc("POST /{$}", h.handleCreate)

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		return fmt.Errorf("serve HTTP: %w", err)
	}

	return nil
}

func (h handler) handleIndex(w http.ResponseWriter, req *http.Request) {
	drops, err := h.r.FetchAll(req.Context())
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var body bytes.Buffer
	err = h.tmpl.ExecuteTemplate(&body, "index.html.tmpl", drops)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = body.WriteTo(w)
}

func (h handler) handleCreate(w http.ResponseWriter, req *http.Request) {
	text := req.FormValue("text")
	if text == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err := h.r.Save(req.Context(), text)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, req, "/", http.StatusSeeOther)
}
