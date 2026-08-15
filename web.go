package main

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/redis/go-redis/v9"
)

func initWeb(redisClient *redis.Client) error {
	tmpl, err := template.ParseGlob("templates/*")
	if err != nil {
		return err
	}

	return http.ListenAndServe(":8080", newHandler(redisClient, tmpl))
}

func newHandler(redisClient *redis.Client, tmpl *template.Template) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		drops, err := fetchAllDrops(r.Context(), redisClient)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		var body bytes.Buffer
		err = tmpl.ExecuteTemplate(&body, "index.html.tmpl", drops)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = body.WriteTo(w)
	})
	mux.HandleFunc("POST /{$}", func(w http.ResponseWriter, r *http.Request) {
		text := r.FormValue("text")
		if text == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		err := saveDrop(r.Context(), redisClient, text)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
	return mux
}
