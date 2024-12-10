package main

import (
		"net/http"
		"html/template"
		"fmt"
)

func handlefunc(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("index.html")
	tmpl.Execute(w, nil)
}

func handlelog(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1> welcome !</h1>")
}

func auth(handler http.Handler) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		if r.FormValue("username") == "" || r.FormValue("password") == "" {
			fmt.Fprintf(w, "<h1>error empty string!</h1>")
			return;
		}
		handler.ServeHTTP(w, r)
	})
}

func main () {
	http.HandleFunc("/", handlefunc)
	http.Handle("/log", auth(http.HandlerFunc(handlelog)))
	http.ListenAndServe(":8080", nil)
}
