package godoc2md

import (
	"bytes"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestURLForFile(t *testing.T) {
	u := urlForFile("github.com/miekg/dns/scan.go", "github.com/miekg/dns", "main")
	if exp := "https://github.com/miekg/dns/blob/main/scan.go"; u != exp {
		t.Errorf("expected %s, got %s", exp, u)
	}

	u = urlForFile("github.com/miekg/gitlabutil/scan.go", "github.com/miekg/gitlabutil", "main")
	if exp := "https://github.com/miekg/gitlabutil/blob/main/scan.go"; u != exp {
		t.Errorf("expected %s, got %s", exp, u)
	}
	u = urlForFile("gitlab.com/miekg/dns/scan.go", "gitlab.com/miekg/dns", "main")
	if exp := "https://gitlab.com/miekg/dns/-/blob/main/scan.go"; u != exp {
		t.Errorf("expected %s, got %s", exp, u)
	}
}

func TestGoDoc(t *testing.T) {
	config := &Config{Import: "testdata"}

	buf := &bytes.Buffer{}
	err := Transform(buf, "testdata", config)
	if err != nil {
		t.Fatal(err)
	}
	exp, err := os.ReadFile("testdata/testdata.md")
	if err != nil {
		t.Fatal(err)
	}
	got := bytes.TrimSpace(buf.Bytes())
	got = append(got, '\n')
	exp = bytes.Replace(exp, []byte("\n\n"), []byte("\n"), -1)
	got = bytes.Replace(got, []byte("\n\n"), []byte("\n"), -1)
	got = bytes.Replace(got, []byte(" \n"), []byte("\n"), -1) // add space newline is something inserted.
	diff := cmp.Diff(exp, got)
	if diff != "" {
		t.Errorf("unexpected diff: %s", diff)
	}
}
