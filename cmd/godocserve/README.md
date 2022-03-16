# Godocserve

Godocserve will take a file containing Go repositories links (one per line). It will create a simple
webserver that allows to show and search for all the documenation. Similar to go.pkg.dev it prefixes
the doc with the README.md - if found in that repo.

Specifically it

1. downloads (--depth 1) each repo (5 in parallel) in the go generate setup
2. gather the docs for this repo (also the subpackages(!))
3. capture the result in /path/to/repo/README.md
4. Index the contents for search
5. Use that path to serve HTML for the docs (via mmark)

It features an index and search page and will display all godoc generated from the downloaded repos.
go:embed is used to add all the files to the binary, so you can just copy it around and not worry
about the files on disk.

## Repos

Each repo path should be an URL that contains the Go code, this should also be the import path.

Optionally you can specify a git branch separated by white space on the same line as well. If not
given it defaults to 'main'. And further more if you need a vanity import you can specify this after
the branch. This does make the branch mandatory to be specified.

By default this files should be named 'repos' (as this is used in the go generate line).

`#` can be used to signal a comment in the file itself.

## Endpoints

There are two endpoints on this web server:

1. / search and index. Shows search box and an index of all indexed repos.
   If something is searched, the listed repos have only that keyword in them.
2. g/ rendered contents of a package. I.e. g/github.com/miekg/dns shows in the contents
  of the docs of that packages.

## Usage

1. amend repos to your liking
2. go generate
3. go build
4. ./godocserve

## Adjusting

Just edit assets/head.tmpl and the other files to your liking and recompile.
