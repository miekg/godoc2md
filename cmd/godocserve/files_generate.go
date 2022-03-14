//+build ignore

// files_generate.go runs with go generate to retrieve all repos and create the files
// structure that is then embedded in the go binary.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/miekg/godoc2md"
)

var (
	flgParallel = flag.Int("p", 5, "run this many goroutines in parallel")
	flgBranch   = flag.String("b", "main", "default branch to use")
)

func main() {
	flag.Parse()
	if len(flag.Args()) != 1 {
		log.Fatal("Need at least a file to read from")
	}
	repof, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	repos := bytes.Split(repof, []byte{'\n'})

	// Override this function to add a link to the subdir docs.
	// Also see linkify.
	godoc2md.Funcs["subdir_format"] = func(s string) string { return `<a href="/g/` + s + `">` + path.Base(s) + `</a>` }

	var wg sync.WaitGroup
	sem := make(chan int, *flgParallel)
	for i, r := range repos {
		if len(r) == 0 { // last line
			continue
		}
		if bytes.HasPrefix(r, []byte("#")) { // comment
			continue
		}

		log.Printf("%q, being looked at, as %d of %d", r, i+1, len(repos)-1)

		branch := *flgBranch
		rs := bytes.Fields(r)
		repo := string(rs[0])
		if len(rs) == 2 {
			branch = string(rs[1])
		}

		wg.Add(1)
		sem <- 1
		go func() {
			defer func() { <-sem; wg.Done() }()

			if err := transform(repo, branch); err != nil {
				log.Printf("%q, failed to clone: %v", repo, err)
				return
			}
		}()
	}
	wg.Wait()

}

func mkdirAll(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0777)
		if err != nil {
			return err
		}
	}
	return nil
}

// transform clones the repo and writes the documentation markdown to the correct directory.
func transform(repo, branch string) error {
	tmpdir, err := os.MkdirTemp("/tmp", "godocserve")
	if err != nil {
		return err
	}
	defer func() { os.RemoveAll(tmpdir) }()

	git := exec.Command("git", "clone", "--depth", "1", "--branch", branch, repo, tmpdir)
	git.Dir = tmpdir

	out, err := git.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Failed to run git command %q: %v, output %s", git, err, out)
	}

	url, err := url.Parse(repo)
	if err != nil {
		return err
	}
	imp := path.Join(url.Host, url.Path)
	log.Printf("%q, cloned succesfully, with import %q in %q", repo, imp, tmpdir)

	config := &godoc2md.Config{
		DeclLinks:         true,
		Import:            imp,
		GitRef:            branch,
		Replace:           tmpdir,
		SrcLinkHashFormat: "#L%d",
	}

	err = filepath.Walk(tmpdir,
		func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(tmpdir, p)
			if len(rel) > 2 && strings.HasPrefix(rel, ".") {
				return nil
			}
			imptmp := config.Import
			defer func() { config.Import = imptmp; config.SubPackage = "" }()
			if rel != "" && rel != "." {
				if !checkForGoFiles(p) { // no go files, skip
					log.Printf("%q, no Go files in %s, skipping", repo, p)
					return nil
				}
				config.Import += "/" + rel
				config.SubPackage = rel
			}

			gobuf := &bytes.Buffer{}
			err = godoc2md.Transform(gobuf, p, config)
			if err != nil {
				log.Printf("%q, failed to generate markdown", repo)
				return nil
			}

			rbuf := &bytes.Buffer{}
			// If there is a README.md add that too, under a # README section, the docs will then follow under a # Documentation section.
			if readmebuf, err := os.ReadFile(path.Join(p, "README.md")); err == nil {
				rbuf.WriteString("# README\n\n")

				if gobuf.Len() > 10 { // there is go code docs, link to that.
					rbuf.WriteString("[Package documentation](#documentation)\n\n")
				}
				rbuf.Write(readmebuf)
			}

			if gobuf.Len() < 10 && rbuf.Len() == 0 { // bit of a cop out, but this means "no docs found", only return if also no readme
				return nil
			}

			// assemble it all
			buf := &bytes.Buffer{}
			buf.Write(rbuf.Bytes())
			buf.WriteString("\n# Documentation\n\n")
			buf.Write(gobuf.Bytes())

			// Create output.
			readme := path.Join(config.Import, "README.md")
			readme = path.Join("content", readme)
			if err := mkdirAll(path.Dir(readme)); err != nil {
				log.Printf("%q, failed to create containing directory %q, for %s: %v", repo, path.Dir(readme), err)
			}

			if err := os.WriteFile(readme, buf.Bytes(), 0666); err != nil {
				log.Printf("%q, failed to write markdown %q, for %s: %v", repo, readme, err)
			}
			log.Printf("%q, wrote markdown into %q", repo, readme)
			return nil

		})
	return err
}

// checkForGoFiles returns true when there are files in p with a .go extension.
func checkForGoFiles(p string) bool {
	des, err := os.ReadDir(p)
	if err != nil {
		return false
	}
	for _, de := range des {
		if de.IsDir() {
			continue
		}
		if ext := path.Ext(de.Name()); ext == ".go" {
			return true
		}
	}
	return false
}
