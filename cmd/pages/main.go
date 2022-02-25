package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"sync"

	"github.com/miekg/godoc2md"
	"github.com/yuin/goldmark"
)

var (
	flgParallel = flag.Int("p", 5, "run this many goroutines in parallel")
	flgBranch   = flag.String("b", "main", "default branch to use")
	flgZip      = flag.Bool("z", false, "target is zip file instead of directory")
	flgHtml     = flag.Bool("h", false, "transform markdown in HTML before writing")
)

func main() {
	flag.Parse()
	if len(flag.Args()) != 2 {
		log.Fatal("Need at least a file to read from and a directory to output to")
	}
	repof, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	if !*flgZip {
		if err := mkdirAll(flag.Arg(1)); err != nil {
			log.Fatal(err)
		}
	}

	repos := bytes.Split(repof, []byte{'\n'})

	ziprw := &zip.Writer{}
	if *flgZip {
		archive, err := os.Create(flag.Arg(1))
		if err != nil {
			log.Fatal(err)
		}
		ziprw = zip.NewWriter(archive)
		defer archive.Close()
		defer ziprw.Close()
	}

	var wg sync.WaitGroup
	var zipmu sync.Mutex
	sem := make(chan int, *flgParallel)
	for i, r := range repos {
		if len(r) == 0 { // last line
			continue
		}
		log.Printf("Looking at %d, %s", i, r)

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

			buf, err := clone(repo, branch)
			if err != nil {
				log.Printf("Failed to clone repo: %s: %v", repo, err)
				return
			}
			if *flgHtml {
				html := &bytes.Buffer{}
				if err := goldmark.Convert(buf, html); err != nil {
					log.Printf("Not good %s: %v", repo, err)
					return
				}
				buf = html.Bytes()
			}
			url, _ := url.Parse(repo) // parsed in clone() as well
			readme := path.Join(path.Join(url.Host, url.Path), "README.md")
			if *flgHtml {
				readme = path.Join(path.Join(url.Host, url.Path), "README.html")
			}
			if !*flgZip {
				readme = path.Join(flag.Arg(1), readme)
				if err := mkdirAll(path.Dir(readme)); err != nil {
					log.Printf("Not good %s: %v", repo, err)
					return
				}
				log.Printf("Writing to markdown to %s", readme)
				if err := os.WriteFile(readme, buf, 0600); err != nil {
					log.Printf("Not good %s: %v", repo, err)
					return
				}
				return
			}
			// zipper stuff, protected by zipmu, as create/write must be done in a single step.
			zipmu.Lock()
			w, err := ziprw.Create(readme)
			if err != nil {
				log.Printf("Failed to create file %q in zip, for %s: %v", readme, repo, err)
			}
			if _, err := w.Write(buf); err != nil {
				log.Printf("Failed to write markdown %q to zip, for %s: %v", readme, repo, err)
			}
			zipmu.Unlock()
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

// clone clones the repo and returns the generated config or an error. If err is not nil
// the returned buffer contains the output of the command that ran.
func clone(repo, branch string) ([]byte, error) {
	tmpdir, err := os.MkdirTemp("/tmp", "pages")
	if err != nil {
		return nil, err
	}
	defer func() { os.RemoveAll(tmpdir) }()

	git := exec.Command("git", "clone", "--depth", "1", "--branch", branch, repo, tmpdir)
	git.Dir = tmpdir

	out, err := git.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("Failed to run git command %q: %v", git, err)
	}

	url, err := url.Parse(repo)
	if err != nil {
		return nil, err
	}
	imp := path.Join(url.Host, url.Path)
	log.Printf("Working on repo %q, with import %q in %q", repo, imp, tmpdir)

	config := &godoc2md.Config{
		DeclLinks: true,
		Import:    imp,
		GitRef:    branch,
	}

	buf := &bytes.Buffer{}
	err = godoc2md.Transform(buf, tmpdir, config)
	return buf.Bytes(), err
}
