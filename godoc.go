// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package godoc2md creates a markdown representation of a package's godoc.
//
// This is forked from https://github.com/davecheney/godoc2md.  The primary difference being that this version is
// a library that can be used by other packages.
package godoc2md

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"strings"
	"text/template"

	"golang.org/x/tools/godoc"
	"golang.org/x/tools/godoc/vfs"
)

var (
	pres *godoc.Presentation
	fs   = vfs.NameSpace{}

	// Funcs contains the functions used in the template. Of these only subdir_format might be
	// of interest to callers.
	Funcs = map[string]interface{}{
		"comment_md":    commentMdFunc,
		"base":          path.Base,
		"md":            mdFunc,
		"pre":           preFunc,
		"kebab":         kebabFunc,
		"bitscape":      bitscapeFunc, // Escape [] for bitbucket confusion
		"subdir_format": path.Base,
	}
)

//Config contains config options for Godoc2md
type Config struct {
	SrcLinkHashFormat string
	SrcLinkFormat     string
	ShowTimestamps    bool
	DeclLinks         bool
	Verbose           bool
	Replace           string
	Import            string
	GitRef            string // commit, tag, or branch of the repo.
}

func commentMdFunc(comment string) string {
	var buf bytes.Buffer
	toMd(&buf, comment)
	return buf.String()
}

func mdFunc(text string) string {
	text = strings.Replace(text, "*", "\\*", -1)
	text = strings.Replace(text, "_", "\\_", -1)
	return text
}

func preFunc(text string) string {
	return "``` go\n" + text + "\n```"
}

// Removed code line that always substracted 10 from the value of `line`.
// Made format for the source link hash configurable to support source control platforms other than Github.
// Original Source https://github.com/golang/tools/blob/master/godoc/godoc.go#L540
func genSrcPosLinkFunc(srcLinkFormat, srcLinkHashFormat string, config *Config) func(s string, line, low, high int) string {
	return func(s string, line, low, high int) string {
		if srcLinkFormat != "" {
			return fmt.Sprintf(srcLinkFormat, s, line, low, high)
		}

		var buf bytes.Buffer
		template.HTMLEscape(&buf, []byte(s))
		// selection ranges are of form "s=low:high"
		if low < high {
			fmt.Fprintf(&buf, "?s=%d:%d", low, high) // no need for URL escaping
			if line < 1 {
				line = 1
			}
		}
		// line id's in html-printed source are of the
		// form "L%d" (on Github) where %d stands for the line number
		if line > 0 {
			fmt.Fprintf(&buf, srcLinkHashFormat, line) // no need for URL escaping
		}
		b := buf.String()
		if strings.HasPrefix(b, config.Replace) {
			url := urlForFile(b[len(config.Replace):], config.Import, config.GitRef)
			return url
		}
		return b
	}
}

func readTemplate(name, data string) (*template.Template, error) {
	t, err := template.New(name).Funcs(pres.FuncMap()).Funcs(Funcs).Parse(data)
	return t, err
}

func kebabFunc(text string) string {
	s := strings.Replace(strings.ToLower(text), " ", "-", -1)
	s = strings.Replace(s, ".", "-", -1)
	s = strings.Replace(s, "\\*", "42", -1)
	return s
}

func bitscapeFunc(text string) string {
	s := strings.Replace(text, "[", "\\[", -1)
	s = strings.Replace(s, "]", "\\]", -1)
	return s
}

// Transform turns your godoc into markdown.The imp (import) path will be used
// for the generated import statement, the same string is also used for generating
// file 'files' links, but then it will be prefixed with 'https://'.
func Transform(out io.Writer, path string, config *Config) error {
	if config.GitRef == "" {
		config.GitRef = "master" // main??
	}

	corpus := godoc.NewCorpus(fs)
	corpus.Verbose = config.Verbose
	pres = godoc.NewPresentation(corpus)
	pres.TabWidth = 4
	pres.ShowTimestamps = config.ShowTimestamps
	pres.DeclLinks = config.DeclLinks
	pres.URLForSrcPos = genSrcPosLinkFunc(config.SrcLinkFormat, config.SrcLinkHashFormat, config)
	pres.URLForSrc = func(s string) string {
		return urlForFile(s, config.Import, config.GitRef)
	}

	tmpl, err := readTemplate("package.txt", pkgTemplate)
	if err != nil {
		return err
	}

	return write(out, fs, pres, tmpl, path, config.Import)
}

// urlForFile takes path, imp and git ref and sep and creates a link to a file in
// github or gitlab.
func urlForFile(s, imp, ref string) string {
	// We get a string that is the import path, github.com/miekg/dns, from which we need to create
	// an url in the form: https://github.com/miekg/dns/blob/dcb0117c0a48f73fec66233f04a798bd1beb122f/AUTHORS
	// in case of github, or
	// https://gitlab.com/miekg/dns/-/blob/dcb0117c0a48f73fec66233f04a798bd1beb122f/AUTHORS
	// in case of gitlab.
	path := strings.TrimPrefix(s, imp)
	sep := seperatorForHub(s)
	return "https://" + imp + sep + "/blob/" + ref + path
}

// seperatorForHub returns "/-/" or the empty string, if the string s contain gitlab or not.
func seperatorForHub(s string) string {
	slash := strings.Index(s, "/")
	if slash == 0 {
		slash = len(s)
	}
	if strings.Contains(s[:slash], "gitlab") {
		return "/-"
	}
	return ""
}
