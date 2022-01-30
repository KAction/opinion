package main

import (
	"github.com/isbm/textwrap"
	"flag"
	"fmt"
	"log"
	"bytes"
	"net/http"
	"errors"
	"strings"
	"io"
	"encoding/json"
	src "git.sr.ht/~kaction/opinion/src"
	"os"
)

var fetchQuery string = `
query(
	$owner:String!,
	$name:String!,
	$number:Int!,
	$batchSize:Int!,
	$cursor:String
) {
	repository(owner:$owner, name:$name) {
		issue(number: $number) {
			id
			title
			body
			closed
			createdAt
			viewerCanUpdate
			author { login }
			comments(first: $batchSize, after: $cursor) {
				pageInfo {
					endCursor
					hasNextPage
				}
				nodes {
					id
					body
					createdAt
					viewerCanDelete
					viewerCanUpdate
					author { login }
				}
			}
		}
	}
}
`

type Author struct {
	Login string
}

type Comment struct {
	Id string
	Body string
	CreatedAt string
	Author Author

	ViewerCanDelete bool
	ViewerCanUpdate bool
}

type PageInfo struct {
	EndCursor *string
	HasNextPage bool
}

type Comments struct {
	PageInfo PageInfo
	Nodes []Comment
}

type Issue struct {
	Ref src.IssueRef
	Id string
	Author Author
	Title string
	Body string
	Closed bool
	CreatedAt string
	ViewerCanUpdate bool
	Comments Comments
}

type Args struct {
	Ref src.IssueRef
	Token string
}

func parseArgs() (*Args, error) {
	envvar := "GITHUB_TOKEN"
	var ref src.IssueRef

	token, ok := os.LookupEnv(envvar)
	if !ok {
		msg := fmt.Sprintf("Environment variable `%s' is not set",
			envvar)
		return nil, errors.New(msg)
	}
	flag.Var(&ref, "get", "issue reference in owner/repo#nnn format")
	flag.Parse()

	if ref.Number == 0 {
		msg := "Processing data from stdin not implemented"
		return nil, errors.New(msg)
	}
	return &Args{Ref: ref, Token: token}, nil
}

func MustMarshal(value interface{}) []byte {
	bytes, err := json.Marshal(value)
	if err != nil {
		format := "Failed to marshal json value: %v %v"
		panic(fmt.Sprintf(format, value, err))
	}
	return bytes
}

func loadIssue(c *http.Client, ref src.IssueRef, token string) (*Issue, error){
	const endpoint = "https://api.github.com/graphql"
	var issue *Issue
	comments := make([]Comment, 0)

	type Repository struct { Issue *Issue }
	type Data struct {
		Data *struct {
			Repository struct {
				Issue *Issue
			}
		}
	}

	var cursor *string
	if c == nil {
		c = &http.Client{}
	}

	for {
		variables := map[string]interface{} {
			"owner": ref.Owner,
			"name": ref.Repo,
			"number": ref.Number,
			// Can't be bigger: arbitrary limit of GitHub.
			"batchSize": 100,
			"cursor": cursor,
		}
		payload := map[string]interface{} {
			"query": fetchQuery,
			"variables": variables,
		}
		payloadBytes := MustMarshal(payload)

		req, err := http.NewRequest("POST", endpoint,
			bytes.NewReader(payloadBytes))
		if err != nil {
			format := "Failed to create request object: %w"
			return nil, fmt.Errorf(format, err)
		}
		req.Header.Add("Authorization",
			fmt.Sprintf("Bearer %s", token))

		resp, err := c.Do(req)
		if err != nil {
			return nil, fmt.Errorf("Request failed: %w", err)
		}
		defer resp.Body.Close()
		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			format := "Failed to read response body: %w"
			return nil, fmt.Errorf(format, err)
		}
		var value Data

		err = json.Unmarshal(bytes, &value)
		if err != nil {
			format := "Response is not valid json: %w"
			return nil, fmt.Errorf(format, err)
		}
		issue = value.Data.Repository.Issue
		comments = append(comments, issue.Comments.Nodes...)
		if !issue.Comments.PageInfo.HasNextPage {
			break
		}
		cursor = issue.Comments.PageInfo.EndCursor
	}
	issue.Ref = ref
	issue.Comments.Nodes = comments
	return issue, nil
}

func printAttrubution(login string, timestamp string) {
	attribution := fmt.Sprintf(" @%s at %s ", login, timestamp)
	fmt.Printf("  <!--%s%s--->\n\n",
		strings.Repeat("-", 69 - len(attribution)),
		attribution)
}

func printIssue(i *Issue) {
	wrapper := textwrap.NewTextWrap().SetWidth(79)
	status := "(open)"
	if i.Closed {
		status = "(closed)"
	}

	header := fmt.Sprintf("%s/%s#%d%s %s", i.Ref.Owner,
		i.Ref.Repo, i.Ref.Number, status, i.Title)
	fmt.Printf("%s\n", header)
	fmt.Printf("%s\n\n", strings.Repeat("=", len(header)))
	fmt.Printf("%s\n", i.Body)
	printAttrubution(i.Author.Login, i.CreatedAt)

	for _, c := range i.Comments.Nodes {
		fmt.Printf("%s\n", wrapper.Fill(c.Body))
		printAttrubution(c.Author.Login, c.CreatedAt)
	}
}

func main() {
	args, err := parseArgs()
	if err != nil {
		log.Fatal("Failed to parse arguments: %v", err)
	}
	issue, err := loadIssue(nil, args.Ref, args.Token)
	if err != nil {
		log.Fatal("Failed to load issue: %v", err)
	}
	printIssue(issue)
}
