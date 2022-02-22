package main

import (
	"bytes"
	"flag"
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
	if err := mkdirAll(flag.Arg(1)); err != nil {
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

		branch := "master"
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
				log.Printf("Not good: %v", err)
				return
			}
			url, _ := url.Parse(repo) // parsed in clone() as well
			dir := path.Join(flag.Arg(1), path.Join(url.Host, url.Path))
			out := path.Join(dir, "README.md")
			if err := mkdirAll(dir); err != nil {
				log.Printf("Not good %s: %v", repo, err)
				return
			}
			log.Printf("Writing to markdown to %s", out)
			if err := os.WriteFile(out, buf, 0600); err != nil {
				log.Printf("Not good %s: %v", repo, err)
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
		return out, err
	}

	url, err := url.Parse(repo)
	if err != nil {
		return nil, err
	}
	imp := path.Join(url.Host, url.Path)
	log.Printf("Working on repo %q, with import %q in %q", repo, imp, tmpdir)

	config := &godoc2md.Config{
		DeclLinks: true,
		//		Replace:           *flgReplace,
		Import: imp,
		GitRef: "main", // from file
	}

	buf := &bytes.Buffer{}
	err = godoc2md.Transform(buf, tmpdir, config)
	return buf.Bytes(), err
}
