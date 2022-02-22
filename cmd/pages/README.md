# Pages

Pages will take a file containing Go repositories links (one per line), and a directory path.

It will:

1. download (--depth 1) each repo
2. run gogo2doc on the repo
3. capture the result in /path/path/to/repo/README.md
4. delete the repo

Then as a separate step the README.md can be converted to HTML.

## Repos

Each repo path should be an URL that contains the Go code, this should also be the import path.

Optionally you can specify a git branch seperated by white space on the same line as well. If not
given it defaults to 'main'.
