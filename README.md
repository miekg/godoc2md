# godoc2md

This is forked from <a
href="https://github.com/davecheney/godoc2md">https://github.com/davecheney/godoc2md</a>.  The
primary difference being that this version is a library that can be used by other packages.

This a fork of a fork <https://github.com/WillAbides/godoc2md>. But cleanup and simplified.
Point it to a Go repo on disk and generate markdown from it.

I.e. assuming `miekg/dns` is cloned in /tmp/dns, you can run:

~~~
cmd/godoc2md/godoc2md -replace '/tmp/dns:https://github.com/miekg/dns' -import github.com/miekg/dns /tmp/dns
~~~

There is some repeating in the command line, which can be reduced, but it's flexible when you using
vanity URLs for your code for instance.

Note: `godoc2md` is a small cmd line that wrap this library.

## Bugs

Examples are not rendered.
