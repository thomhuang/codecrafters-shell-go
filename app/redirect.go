package main

import "os"

type RedirectType int

const (
	STDOUT RedirectType = iota
	STDERR
)

type Redirect struct {
	file *os.File
	typ  RedirectType
}

type redirectSpec struct {
	flags int
	typ   RedirectType
}

var redirectOps = map[string]redirectSpec{
	">":   {os.O_CREATE | os.O_WRONLY | os.O_TRUNC, STDOUT},  // stdout redirection (overwrite)
	"1>":  {os.O_CREATE | os.O_WRONLY | os.O_TRUNC, STDOUT},  // ^ same as ">"
	">>":  {os.O_CREATE | os.O_WRONLY | os.O_APPEND, STDOUT}, // stdout redirection (append)
	"1>>": {os.O_CREATE | os.O_WRONLY | os.O_APPEND, STDOUT}, // ^ same as ">>"
	"2>":  {os.O_CREATE | os.O_WRONLY | os.O_TRUNC, STDERR},  // stderr redirection (overwrite)
	"2>>": {os.O_CREATE | os.O_WRONLY | os.O_APPEND, STDERR}, // stderr redirection (append)
}

// extractRedirects separates redirect operators (and their targets) from the
// command arguments, opening the target file for the last redirect of each
// stream. The remaining (non-redirect) args are returned alongside.
func extractRedirects(args []string) ([]string, *Redirect) {
	var cleanArgs []string
	var redirect *Redirect

	for i := 0; i < len(args); i++ {
		spec, isRedirect := redirectOps[args[i]]
		if !isRedirect || i+1 >= len(args) { // if not redirect or EOC, treat as literal arg
			cleanArgs = append(cleanArgs, args[i]) // literal arg or dangling operator
			continue
		}

		i++ // skip to the redirect target
		file, err := os.OpenFile(args[i], spec.flags, 0644)

		// if there are multiple redirect operators for the same stream, we only keep the last one
		if redirect != nil {
			redirect.file.Close()
			redirect = nil
		}
		if err != nil { // open failed; drop the redirect
			continue
		}

		redirect = &Redirect{file: file, typ: spec.typ}
	}

	return cleanArgs, redirect
}
