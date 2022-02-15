package godoc2md

import (
	"fmt"
	"go/ast"
	"go/build"
	"io"
	"log"
	"os"
	pathpkg "path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"golang.org/x/tools/godoc"
	"golang.org/x/tools/godoc/vfs"
)

const target = "/target"

// write writes the godoc in pres to w.
func write(w io.Writer, fs vfs.NameSpace, pres *godoc.Presentation, tmpl *template.Template, args []string) error {
	path := args[0]

	abspath, relpath := paths(fs, pres, path)

	info := pres.GetPkgPageInfo(abspath, relpath, 0)

	fmt.Fprintf(os.Stderr, "%+v\n", info.PDoc)

	if info == nil {
		return fmt.Errorf("%s: no such directory or package", args[0])
	}
	if info.Err != nil {
		return info.Err
	}

	source := pres.URLForSrc(".")
	println(source)

	println(info.PDoc.ImportPath)
	if info.PDoc != nil && info.PDoc.ImportPath == target {
		// Replace virtual /target with actual argument from command line.
		info.PDoc.ImportPath = args[0]
	}

	// If we have more than one argument, use the remaining arguments for filtering.
	if len(args) > 1 {
		info.IsFiltered = true
		filterInfo(args[1:], info)
	}

	if err := tmpl.Execute(w, info); err != nil {
		return err
	}

	return nil
}

// paths determines the paths to use.
//
// If we are passed an operating system path like . or ./foo or /foo/bar or c:\mysrc,
// we need to map that path somewhere in the fs name space so that routines
// like getPageInfo will see it.  We use the arbitrarily-chosen virtual path "/target"
// for this.  That is, if we get passed a directory like the above, we map that
// directory so that getPageInfo sees it as /target.
// Returns the absolute and relative paths.
func paths(fs vfs.NameSpace, pres *godoc.Presentation, path string) (abspath, relpath string) {
	if filepath.IsAbs(path) {
		fs.Bind(target, vfs.OS(path), "/", vfs.BindReplace)
		return target, target
	}
	if build.IsLocalImport(path) {
		cwd, err := os.Getwd()
		if err != nil {
			log.Printf("error while getting working directory: %v", err)
		}
		path = filepath.Join(cwd, path)
		fs.Bind(target, vfs.OS(path), "/", vfs.BindReplace)
		return target, target
	}
	bp, err := build.Import(path, "", build.FindOnly)
	if err != nil {
		log.Printf("error while importing build package: %v", err)
	}
	if bp.Dir != "" && bp.ImportPath != "" {
		fs.Bind(target, vfs.OS(bp.Dir), "/", vfs.BindReplace)
		return target, bp.ImportPath
	}
	return pathpkg.Join(pres.PkgFSRoot(), path), path
}

// filterInfo updates info to include only the nodes that match the given
// filter args.
func filterInfo(args []string, info *godoc.PageInfo) {
	rx, err := makeRx(args)
	if err != nil {
		log.Fatalf("illegal regular expression from %v: %v", args, err)
	}

	filter := func(s string) bool {
		fmt.Fprintf(os.Stderr, "filtering on string: %s\n", s)

		return rx.MatchString(s)
	}
	switch {
	case info.PAst != nil:
		newPAst := map[string]*ast.File{}
		for name, a := range info.PAst {
			cmap := ast.NewCommentMap(info.FSet, a, a.Comments)
			a.Comments = []*ast.CommentGroup{} // remove all comments.
			ast.FilterFile(a, filter)
			if len(a.Decls) > 0 {
				newPAst[name] = a
			}
			for _, d := range a.Decls {
				// add back the comments associated with d only
				comments := cmap.Filter(d).Comments()
				a.Comments = append(a.Comments, comments...)
			}
		}
		info.PAst = newPAst // add only matching files.
	case info.PDoc != nil:
		info.PDoc.Filter(filter)
	}
}

// Does s look like a regular expression?
func isRegexp(s string) bool {
	return strings.ContainsAny(s, ".(|)*+?^$[]")
}

// Make a regular expression of the form
// names[0]|names[1]|...names[len(names)-1].
// Returns an error if the regular expression is illegal.
func makeRx(names []string) (*regexp.Regexp, error) {
	if len(names) == 0 {
		return nil, fmt.Errorf("no expression provided")
	}
	s := ""
	for i, name := range names {
		if i > 0 {
			s += "|"
		}
		if isRegexp(name) {
			s += name
		} else {
			s += "^" + name + "$" // must match exactly
		}
	}
	fmt.Fprintf(os.Stderr, "regex string: %s\n", s)
	return regexp.Compile(s)
}
