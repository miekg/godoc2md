// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package godoc2md creates a markdown representation of a package's godoc.
//
// This is forked from https://github.com/davecheney/godoc2md.  The primary difference being that this version is
// a library that can be used by other packages.
package godoc2md

//go:generate bin/goreadme github.com/WillAbides/godoc2md

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

	funcs = map[string]interface{}{
		"comment_md":  commentMdFunc,
		"base":        path.Base,
		"md":          mdFunc,
		"pre":         preFunc,
		"kebab":       kebabFunc,
		"bitscape":    bitscapeFunc, //Escape [] for bitbucket confusion
		"trim_prefix": strings.TrimPrefix,
	}
)

//Config contains config options for Godoc2md
type Config struct {
	SrcLinkHashFormat string
	SrcLinkFormat     string
	ShowTimestamps    bool
	DeclLinks         bool
	Verbose           bool
	Replace           map[string]string
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
// The replace map will replace any prefix of the generated string with the value of that key.
func genSrcPosLinkFunc(srcLinkFormat, srcLinkHashFormat string, replace map[string]string) func(s string, line, low, high int) string {
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
		for k, v := range replace {
			if strings.HasPrefix(b, k) {
				b = strings.Replace(b, k, v, 1)
				break
			}
		}
		return b
	}
}

func readTemplate(name, data string) (*template.Template, error) {
	// be explicit with errors (for app engine use)
	t, err := template.New(name).Funcs(pres.FuncMap()).Funcs(funcs).Parse(data)
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

// Transform turns your godoc into markdown.
func Transform(out io.Writer, path, imp string, config *Config) error {
	corpus := godoc.NewCorpus(fs)
	corpus.Verbose = config.Verbose
	pres = godoc.NewPresentation(corpus)
	pres.TabWidth = 4
	pres.ShowTimestamps = config.ShowTimestamps
	pres.DeclLinks = config.DeclLinks
	pres.URLForSrcPos = genSrcPosLinkFunc(config.SrcLinkFormat, config.SrcLinkHashFormat, config.Replace)

	tmpl, err := readTemplate("package.txt", pkgTemplate)
	if err != nil {
		return err
	}

	return write(out, fs, pres, tmpl, path, imp)
}
