package main

//go:generate go run files_generate.go repos
import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gorilla/mux"
	"github.com/mmarkdown/mmark/mparser"
)

// TODO: css and other fluff, put in assets/  see below

//go:embed content/* assets/*
var content embed.FS

func main() {
	// fs Walk the dirs and index everything.

	r := mux.NewRouter()
	r.PathPrefix("/assets").Handler(http.FileServer(http.FS(content)))
	r.PathPrefix("/g").HandlerFunc(renderHandler)
	r.PathPrefix("/").HandlerFunc(searchHandler)

	log.Print("Starting up on 8080")
	log.Fatal(http.ListenAndServe(":8080", r))

}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello world!")
}

func renderHandler(w http.ResponseWriter, r *http.Request) {
	p, err := filepath.Abs(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Print("path", p)
	p = pathForReadme(p)

	data, err := content.ReadFile(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	data, err = htmlify(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func pathForReadme(p string) string {
	p = strings.TrimPrefix(p, "/g")
	p = filepath.Join("content", p+"/README.md")
	return p
}

func htmlify(buf []byte) ([]byte, error) {
	title := "todo"
	p := parser.NewWithExtensions(mparser.Extensions)
	doc := markdown.Parse(buf, p)

	// create fragement so we need to inject header and footer stuff from assets dir
	opts := html.RendererOptions{Flags: html.CommonFlags | html.FootnoteNoHRTag | html.FootnoteReturnLinks}
	r := html.NewRenderer(opts)
	buf = markdown.Render(doc, r)

	headtmpl, err := template.ParseFS(content, "assets/head.html")
	if err != nil {
		return nil, err
	}
	foottmpl, _ := template.ParseFS(content, "assets/foot.html")
	if err != nil {
		return nil, err
	}

	headbuf := &bytes.Buffer{}
	headtmpl.Execute(headbuf, title)
	footbuf := &bytes.Buffer{}
	foottmpl.Execute(footbuf, nil)

	buf = append(headbuf.Bytes(), buf...)
	buf = append(buf, footbuf.Bytes()...)

	return buf, nil
}
