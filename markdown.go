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
func write(w io.Writer, fs vfs.NameSpace, pres *godoc.Presentation, tmpl *template.Template, path string) error {

	fs.Bind(path, vfs.OS(path), "/", vfs.BindReplace)
	// the . is the importpath being shown, we need to make something sensible from path
	info := pres.GetPkgPageInfo(path, ".", 0)

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
