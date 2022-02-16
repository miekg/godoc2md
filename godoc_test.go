package godoc2md

import "testing"

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
