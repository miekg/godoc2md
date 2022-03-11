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

	var wg sync.WaitGroup
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
			url, _ := url.Parse(repo) // parsed in clone() as well
			readme := path.Join(path.Join(url.Host, url.Path), "README.md")
			readme = path.Join("content", readme)
			if err := mkdirAll(path.Dir(readme)); err != nil {
				log.Printf("Failed to create containing directory %q, for %s: %v", path.Dir(readme), repo, err)
				return
			}

			if err := os.WriteFile(readme, buf, 0666); err != nil {
				log.Printf("Failed to write markdown %q, for %s: %v", readme, repo, err)
			}
			log.Printf("Wrote markdown into %q, for %s", readme, repo)
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
	tmpdir, err := os.MkdirTemp("/tmp", "godocserve")
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
		DeclLinks:         true,
		Import:            imp,
		GitRef:            branch,
		Replace:           tmpdir,
		SrcLinkHashFormat: "#L%d",
	}

	buf := &bytes.Buffer{}
	err = godoc2md.Transform(buf, tmpdir, config)
	return buf.Bytes(), err
}
