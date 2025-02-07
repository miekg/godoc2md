package main

import "testing"

func TestLinkify(t *testing.T) {
	link := linkify("content/github.com/miekg/dns/README.md")
	if link != `<a href="/g/github.com/miekg/dns">github.com/miekg/dns</a>` {
		t.Errorf("failed to convert link correctly with linkify, got %s", link)
	}
}
