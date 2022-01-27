package main

import (
	"flag"
	"fmt"
	src "git.sr.ht/~kaction/opinion/src"
)

func main() {
	var issueRef src.IssueRef
	flag.Var(&issueRef, "get", "issue reference in owner/repo#nnn format")
	flag.Parse()

	fmt.Printf("%v", issueRef)
}
