package main

//go:generate go run files_generate.go repos
import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"

	bleve "github.com/blevesearch/bleve/v2"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
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
	r.PathPrefix("/assets").Handler(handlers.LoggingHandler(os.Stdout, http.FileServer(http.FS(content))))
	r.PathPrefix("/g").Handler(handlers.LoggingHandler(os.Stdout, http.HandlerFunc(renderHandler)))
	r.PathPrefix("/").Handler(handlers.LoggingHandler(os.Stdout, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := searchContext{
			Index: index,
		}
		s.searchHandler(w, r)
	})))

	log.Printf("Starting up on: :%d", *flgPort)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*flgPort), r))
}
