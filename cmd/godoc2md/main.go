// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// godoc2md converts godoc formatted package documentation into Markdown format.
//
//
// Usage
//
//    godoc2md $PACKAGE > $GOPATH/src/$PACKAGE/README.md
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/miekg/godoc2md"
)

var (
	verbose = flag.Bool("v", false, "verbose mode")

	// layout control
	showTimestamps = flag.Bool("timestamps", false, "show timestamps with directory listings")
	declLinks      = flag.Bool("links", true, "link identifiers to their declarations")

	// The hash format for Github is the default `#L%d`; but other source control platforms do not
	// use the same format. For example Bitbucket Enterprise uses `#%d`. This option provides the
	// user the option to switch the format as needed and still remain backwards compatible.
	srcLinkHashFormat = flag.String("hashformat", "#L%d", "source link URL hash format")

	srcLinkFormat = flag.String("srclink", "", "if set, format for entire source link")

	flgImport  = flag.String("import", "", "import path for the package")
	flgReplace = flag.String("replace", "", "replace package source with import path")
	flgRef     = flag.String("gitref", "master", "git ref to use for generating the files' link")
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: godoc2md [options] package\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 1 {
		usage()
	}
	pkgName := flag.Arg(0) // actually path

	config := &godoc2md.Config{
		ShowTimestamps:    *showTimestamps,
		DeclLinks:         *declLinks,
		SrcLinkHashFormat: *srcLinkHashFormat,
		SrcLinkFormat:     *srcLinkFormat,
		Verbose:           *verbose,
		Replace:           *flgReplace,
		Import:            *flgImport,
		GitRef:            *flgRef,
	}

	err := filepath.Walk(pkgName,
		func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(pkgName, p)
			if len(rel) > 2 && strings.HasPrefix(rel, ".") {
				return nil
			}
			imp := config.Import
			defer func() { config.Import = imp; config.SubPackage = "" }()
			if rel != "" && rel != "." {
				config.Import += "/" + rel
				config.SubPackage = rel
			}

			if err := godoc2md.Transform(os.Stdout, p, config); err != nil {
				log.Println(err)
			}
			return nil
		})
	if err != nil {
		log.Fatal(err)
	}
}
