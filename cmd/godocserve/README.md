# Godocserve

Godocserve will take a file containing Go repositories links (one per line).

It will:

1. download (--depth 1) each repo
2. run gogo2doc on the repo
3. capture the result in /path/path/to/repo/README.md
4. Use that path to serve HTML for the docs.

It features an index and search page and will display all godoc learned from the downloaded repos.

## Repos

Each repo path should be an URL that contains the Go code, this should also be the import path.

Optionally you can specify a git branch seperated by white space on the same line as well. If not
given it defaults to 'main'.

By default this files should be named 'repos'

## Endpoints

There are two endpoints on this webserver:

1. / search and index. Shows search box and an alphabetical index of all indexed repos.
   If something is search the listed repos have only that keyword.
2. g/ rendered contents of a package. I.e. g/github.com/miekg/dns shows in the contents
  of the docs of that packages.
