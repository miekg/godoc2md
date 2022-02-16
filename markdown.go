package godoc2md

import (
	"fmt"
	"io"
	"text/template"

	"golang.org/x/tools/godoc"
	"golang.org/x/tools/godoc/vfs"
)

const target = "/target"

// write writes the godoc in pres to w.
func write(w io.Writer, fs vfs.NameSpace, pres *godoc.Presentation, tmpl *template.Template, path, imp string) error {

	fs.Bind(path, vfs.OS(path), "/", vfs.BindReplace)
	info := pres.GetPkgPageInfo(path, imp, 0)
	/*
		for i := range info.Examples {
			println(info.Examples[i].Name)

		}
	*/

	if info == nil {
		return fmt.Errorf("%s: no such directory or package", path)
	}
	if info.Err != nil {
		return info.Err
	}

	if err := tmpl.Execute(w, info); err != nil {
		return err
	}

	return nil
}
