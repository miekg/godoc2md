package main

import "testing"

func TestLinkify(t *testing.T) {
	link := linkify("content/github.com/miekg/dns/README.md")
	if link != `<a href="http://localhost:8080/g/github.com/miekg/dns/README.md">github.com/miekg/dns/README.md</a>` {
		t.Errorf("failed to convert link correctly with linkify, got %s", link)
	}
}
