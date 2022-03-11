# godoc2md

This is forked from <a
href="https://github.com/davecheney/godoc2md">https://github.com/davecheney/godoc2md</a>.  The
primary difference being that this version is a library that can be used by other packages.

This a fork of a fork <https://github.com/WillAbides/godoc2md>. But cleanup and simplified.
Point it to a Go repo on disk and generate markdown from it.

I.e. assuming `miekg/dns` is cloned in /tmp/dns, you can run:

~~~
cmd/godoc2md/godoc2md -replace '/tmp/dns' -import 'github.com/miekg/dns' /tmp/dns
~~~

`-replace` removes the "/tmp/dns" prefix from the files in /tmp/dns and allows for the creation
of the correct link to the code using the `-import` path. Optionally the git reference can be
given as well, currently this defaults to "master".

Note: `godoc2md` is a small cmd line that wrap this library. Library usage can be pulled from it.

## Bugs

Examples are not rendered. Notes and Bugs section is either not rendered or done as HTML.
Subdirectories also doesn't look right.

# godocserve

Godocserve is much more interesting as it will render all downloaded markdown as HTML and has
a search option. See cmd/godocserve/README.md for more information.
