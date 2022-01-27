package src
import (
	"errors"
	"regexp"
	"strconv"
	"fmt"
)
var issueRefRegex *regexp.Regexp

type IssueRef struct {
	Owner  string
	Repo   string
	Number int
}

func (v *IssueRef) String() string {
	return fmt.Sprintf("%s/%s#%d", v.Owner, v.Repo, v.Number)
}

func init() {
	// Regular expression is more lax than starndard owner/repo#N
	// because my shell requires quoting of # even in interactive mode.
	s := "^([-a-zA-Z0-9]+).([-a-zA-Z0-9]+).([1-9][0-9]*)$"
	rx, err := regexp.Compile(s)
	if err != nil {
		panic("bug: issueRefRegex is invalid")
	}
	issueRefRegex = rx
}

func (v *IssueRef) Set(s string) error {
	m := issueRefRegex.FindStringSubmatch(s)
	if m == nil {
		return errors.New("no match for `owner/repo#nnn' format")
	}
	number, err := strconv.Atoi(m[3])
	if err != nil {
		panic("issueRefRegex accepted non-numeric issue number")
	}
	v.Owner, v.Repo, v.Number = m[1], m[2], number
	return nil
}
