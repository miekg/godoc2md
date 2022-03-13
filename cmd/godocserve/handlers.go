package main

import (
	"bytes"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	bleve "github.com/blevesearch/bleve/v2"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/mmarkdown/mmark/mparser"
)

func (s searchContext) searchHandler(w http.ResponseWriter, r *http.Request) {
	term := r.FormValue("search")
	var request *bleve.SearchRequest
	if term == "" {
		// TODO: sort these
		query := bleve.NewMatchAllQuery()
		request = bleve.NewSearchRequest(query)
	} else {
		query := bleve.NewQueryStringQuery(term)
		request = bleve.NewSearchRequest(query)
	}

	docs, _ := s.DocCount()
	request.Size = int(docs)
	results, err := s.Search(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	FuncMap := template.FuncMap{
		"linkify": linkify,
	}

	type tmplContext struct {
		*bleve.SearchResult
		Term string
	}

	searchtmpl, err := template.New("search.tmpl").Funcs(FuncMap).ParseFS(content, "assets/search.tmpl")
	if err != nil {
		panic(err)
	}
	searchbuf := &bytes.Buffer{}
	ctx := &tmplContext{
		SearchResult: results,
		Term:         term,
	}
	if err := searchtmpl.Execute(searchbuf, ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buf := append(header("Search Results"), searchbuf.Bytes()...)
	buf = append(buf, footer()...)

	w.Write(buf)
}

func renderHandler(w http.ResponseWriter, r *http.Request) {
	title, err := filepath.Abs(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	title = strings.TrimPrefix(title, "/g/")

	p := pathForReadme(title)
	data, err := content.ReadFile(p)
	if err != nil {
		// check if we find a directory and how that instead
		des, err := os.ReadDir(filepath.Dir(p))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		FuncMap := template.FuncMap{
			"dirify": dirify,
		}
		dirtmpl, err := template.New("directory.tmpl").Funcs(FuncMap).ParseFS(content, "assets/directory.tmpl")
		if err != nil {
			panic(err)
		}
		// massage dir entries to make template easier
		dirs := []string{}
		for _, de := range des {
			if !de.IsDir() {
				continue
			}
			dirs = append(dirs, filepath.Join(filepath.Dir(p), de.Name()))
		}
		dirbuf := &bytes.Buffer{}
		if err := dirtmpl.Execute(dirbuf, dirs); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		buf := append(header(filepath.Dir(p)), dirbuf.Bytes()...)
		buf = append(buf, footer()...)
		w.Write(buf)
		return

	}
	data, err = htmlify(data, title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func pathForReadme(p string) string { return filepath.Join("content", p+"/README.md") }

func htmlify(buf []byte, title string) ([]byte, error) {
	p := parser.NewWithExtensions(mparser.Extensions)
	doc := markdown.Parse(buf, p)

	// create fragement so we need to inject header and footer stuff from assets dir
	opts := html.RendererOptions{Flags: html.CommonFlags | html.FootnoteNoHRTag | html.FootnoteReturnLinks}
	r := html.NewRenderer(opts)
	buf = markdown.Render(doc, r)

	buf = append(header(title), buf...)
	buf = append(buf, footer()...)

	return buf, nil
}

func header(title string) []byte {
	headtmpl, err := template.ParseFS(content, "assets/head.tmpl")
	if err != nil {
		panic(err)
	}
	headbuf := &bytes.Buffer{}
	headtmpl.Execute(headbuf, title)
	return headbuf.Bytes()
}

func footer() []byte {
	foottmpl, err := template.ParseFS(content, "assets/foot.tmpl")
	if err != nil {
		panic(err)
	}
	footbuf := &bytes.Buffer{}
	foottmpl.Execute(footbuf, nil)
	return footbuf.Bytes()
}

// linkify converts "content/github.com/miekg/dns/README.md" in a link
// <a href="/g/github.com/miekg/dns">github.com/miekg/dns</a>.
func linkify(s string) string {
	link := strings.TrimPrefix(s, "content/")
	link = path.Dir(link)
	return `<a href="/g/` + link + `">` + link + `</a>`
}

// dirify converts "content/github.com/miekg/dns/" in a link
// <a href="/g/github.com/miekg/dns">github.com/miekg/dns</a>.
func dirify(s string) string {
	link := strings.TrimPrefix(s, "content/")
	return `<a href="/g/` + link + `">` + link + `</a>`
}
