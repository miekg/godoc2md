package main

//go:generate go run files_generate.go repos
import (
	"bytes"
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	bleve "github.com/blevesearch/bleve/v2"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gorilla/mux"
	"github.com/mmarkdown/mmark/mparser"
)

//go:embed content/* assets/*
var content embed.FS

type searchContext struct {
	bleve.Index
}

var (
	flgPort = flag.Int("p", 8080, "port to listen on")
)

func main() {
	flag.Parse()

	mapping := bleve.NewIndexMapping()
	index, err := bleve.NewMemOnly(mapping)
	if err != nil {
		log.Fatal(err)
	}
	if err := fs.WalkDir(content, "content", func(p string, d fs.DirEntry, walkErr error) error {
		if d.Name() == "README.md" {
			data, err := content.ReadFile(p)
			if err != nil {
				return err
			}
			// We need to convert this data to a string, otherwise things are not indexed correctly.
			if err := index.Index(p, string(data)); err != nil {
				return err

			}
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}
	docs, _ := index.DocCount()
	log.Printf("Indexed %d documents", docs)

	r := mux.NewRouter()
	r.PathPrefix("/assets").Handler(http.FileServer(http.FS(content)))
	r.PathPrefix("/g").HandlerFunc(renderHandler)
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := searchContext{
			Index: index,
		}
		s.searchHandler(w, r)
	})

	log.Printf("Starting up on: :%d", *flgPort)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*flgPort), r))
}

func (s searchContext) searchHandler(w http.ResponseWriter, r *http.Request) {
	term := r.FormValue("search")
	var search *bleve.SearchRequest
	if term == "" {
		query := bleve.NewMatchAllQuery()
		search = bleve.NewSearchRequest(query)
	} else {
		query := bleve.NewQueryStringQuery(term)
		search = bleve.NewSearchRequest(query)
	}

	results, err := s.Search(search)
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

	searchtmpl, err := template.New("search.html").Funcs(FuncMap).ParseFS(content, "assets/search.html")
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
		http.Error(w, err.Error(), http.StatusNotFound)
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
	headtmpl, err := template.ParseFS(content, "assets/head.html")
	if err != nil {
		panic(err)
	}
	headbuf := &bytes.Buffer{}
	headtmpl.Execute(headbuf, title)
	return headbuf.Bytes()
}

func footer() []byte {
	foottmpl, err := template.ParseFS(content, "assets/foot.html")
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
