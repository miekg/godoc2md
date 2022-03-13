package godoc2md

import (
	"fmt"
	"io"
	"text/template"

	"golang.org/x/tools/godoc"
	"golang.org/x/tools/godoc/vfs"
)

// write writes the godoc in pres to w.
func write(w io.Writer, fs vfs.NameSpace, pres *godoc.Presentation, tmpl *template.Template, path, imp string) error {
	fs.Bind(path, vfs.OS(path), "/", vfs.BindReplace) // ??
	info := pres.GetPkgPageInfo(path, imp, 0)

	/*
		for i := range info.Examples {
			fmt.Fprintf(os.Stderr, "** %s\n", info.Examples[i].Name)

			set := token.NewFileSet()
			if info.Examples[i].Play != nil {
				format.Node(os.Stderr, set, info.Examples[i].Play)
			} else {
				format.Node(os.Stderr, set, info.Examples[i].Code)
			}
			println(info.Examples[i].Doc)
			println()
			println()
		}
	*/

	if info == nil {
		return fmt.Errorf("%s: no such directory or package", path)
	}
	if info.Err != nil {
		return info.Err
	}
	return tmpl.Execute(w, info)
}
