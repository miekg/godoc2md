package godoc2md

var pkgTemplate = `{{with .PDoc}}
{{if $.IsMain}}
> {{ base .ImportPath }}
{{comment_md .Doc}}
{{else}}
# {{ .Name }}
` + "`" + `import "{{.ImportPath}}"` + "`" + `

* [Overview](#pkg-overview)
* [Index](#pkg-index){{if $.Examples}}
* [Examples](#pkg-examples){{- end}}{{if $.Dirs}}
* [Subdirectories](#pkg-subdirectories){{- end}}

## Overview {#pkg-overview}
{{comment_md .Doc}}
{{example_html $ ""}}

## Index{{if .Consts}} {#pkg-index}
* [Constants](#pkg-constants){{end}}{{if .Vars}}
* [Variables](#pkg-variables){{end}}{{- range .Funcs -}}{{$name_html := html .Name}}
* [{{node_html $ .Decl false | sanitize}}](#{{$name_html}}){{- end}}{{- range .Types}}{{$tname_html := html .Name}}
* [type {{$tname_html}}](#{{$tname_html}}){{- range .Funcs}}{{$name_html := html .Name}}
  * [{{node_html $ .Decl false | sanitize}}](#{{$name_html}}){{- end}}{{- range .Methods}}{{$name_html := html .Name}}
  * [{{node_html $ .Decl false | sanitize}}](#{{$tname_html}}.{{$name_html}}){{- end}}{{- end}}{{- if $.Notes}}{{- range $marker, $item := $.Notes}}
* [{{noteTitle $marker | html}}s](#pkg-note-{{$marker}}){{end}}{{end}}
{{if $.Examples}}
#### Examples {#pkg-examples} {{- range $.Examples}}
* [{{example_name .Name}}](#example_{{.Name}}){{- end}}{{- end}}
{{with .Filenames}}
#### Package files {#pkg-files}
{{range .}}[{{.|filename|html}}]({{.|srcLink|html}}) {{end}}
{{end}}

{{with .Consts}}## Constants {#pkg-constants}
{{range .}}{{node $ .Decl | pre}}
{{comment_md .Doc}}{{end}}{{end}}
{{with .Vars}}## Variables {#pkg-variables}
{{range .}}{{node $ .Decl | pre}}
{{comment_md .Doc}}{{end}}{{end}}

{{range .Funcs}}{{$name_html := html .Name}}## func {#{{$name_html}}} [{{$name_html}}]({{posLink_url $ .Decl}})
{{node $ .Decl | pre}}
{{comment_md .Doc}}
{{example_html $ .Name}}
{{callgraph_html $ "" .Name}}{{end}}
{{range .Types}}{{$tname := .Name}}{{$tname_html := html .Name}}## type {#{{$tname_html}}} [{{$tname_html}}]({{posLink_url $ .Decl}})
{{node $ .Decl | pre}}
{{comment_md .Doc}}{{range .Consts}}
{{node $ .Decl | pre }}
{{comment_md .Doc}}{{end}}{{range .Vars}}
{{node $ .Decl | pre }}
{{comment_md .Doc}}{{end}}

{{example_html $ $tname}}
{{implements_html $ $tname}}
{{methodset_html $ $tname}}

{{range .Funcs}}{{$name_html := html .Name}}### func {#{{$name_html}}} [{{$name_html}}]({{posLink_url $ .Decl}})
{{node $ .Decl | pre}}
{{comment_md .Doc}}
{{example_html $ .Name}}{{end}}
{{callgraph_html $ "" .Name}}

{{range .Methods}}{{$name_html := html .Name}}### func {#{{$tname_html}}.{{$name_html}}} ({{md .Recv}}) [{{$name_html}}]({{posLink_url $ .Decl}})
{{node $ .Decl | pre}}
{{comment_md .Doc}}
{{$name := printf "%s_%s" $tname .Name}}{{example_html $ $name}}
{{callgraph_html $ .Recv .Name}}
{{end}}{{end}}{{end}}

{{with $.Notes}}
{{range $marker, $content := .}}
## <a name="pkg-note-{{$marker}}">{{noteTitle $marker | html}}s
<ul style="list-style: none; padding: 0;">
{{range .}}
<li><a href="{{posLink_url $ .}}">&#x261e;</a> {{html .Body}}</li>
{{end}}
</ul>
{{end}}
{{end}}
{{end}}
`
