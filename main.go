package main

import (
	"flag"
	"fmt"
	src "git.sr.ht/~kaction/opinion/src"
	"os"
)

func main() {
	var issueRef src.IssueRef
	flag.Var(&issueRef, "get", "issue reference in owner/repo#nnn format")
	flag.Parse()
	if issueRef.Number == 0 {
		fmt.Println("Processing data from stdin not implemented")
		os.Exit(1)
	}

	fmt.Printf("%v", issueRef)
}
