package main

import (
	"flag"
	"html/template"
	"log"
	"path/filepath"
	"sync"

	"net/http"
)

type (
	templateHandler struct {
		once     sync.Once
		filename string
		tmpl     *template.Template
	}
)

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.tmpl = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})

	t.tmpl.Execute(w, r)
}	

func main() {
	var addr = flag.String("addr", ":8080", "http service address")
	flag.Parse()

	r := newRoom()

	http.Handle("/", &templateHandler{filename: "chat.html"})
	http.Handle("/room", r)

	// Run the room in a separate goroutine
	go r.run()

	// Start the web server
	log.Println("Starting web server on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
