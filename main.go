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

type Issue struct {
	Ref src.IssueRef
	Id string
	Title string
	Body string
	Closed bool
	ViewerCanUpdate bool
	Comments []Comment
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


func main() {
	args, err := parseArgs()
	if err != nil {
		log.Fatal("Failed to parse arguments: %v", err)
	}

	var cursor *string
	variables := map[string]interface{} {
		"owner": args.Ref.Owner,
		"name": args.Ref.Repo,
		"number": args.Ref.Number,
		"batchSize": 20,
		"cursor": cursor,
	}
	payload := map[string]interface{} {
		"query": fetchQuery,
		"variables": variables,
	}
	payloadBS, err := json.Marshal(payload)
	if err != nil {
		log.Fatal("Failed to marshal json: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.github.com/graphql",
		bytes.NewReader(payloadBS))
	if err != nil {
		log.Fatal("Failed to create request object: %v", err)
	}
		
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", args.Token))
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Request failed: %v", err)
	}
	defer resp.Body.Close()
	text, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Failed to read response content: %v", err)
	}
	var value map[string]interface{}

	err = json.Unmarshal([]byte(text), &value)
	if err != nil {
		log.Fatal("Failed to parse response json")
	}
	value = value["data"].(map[string]interface{})
	value = value["repository"].(map[string]interface{})
	value = value["issue"].(map[string]interface{})
	createdAt := value["createdAt"].(string)
	author := value["author"].(map[string]interface{})
	closed := value["closed"].(bool)
	title := value["title"].(string)
	body := value["body"].(string)
	authorLogin := author["login"].(string)
	status := "(open)"
	if closed {
		status = "(closed)"
	}
	// Ugly code that formats issue & comments somehow, without
	// unmarshalling json into proper datatype hierarchy. Ugly python
	// style.
	header := fmt.Sprintf("%s/%s#%d%s %s", args.Ref.Owner,
		args.Ref.Repo, args.Ref.Number, status, title)
	fmt.Printf("%s\n", header)
	fmt.Printf("%s\n\n", strings.Repeat("=", len(header)))
	fmt.Printf("%s\n", body)
	attribution := fmt.Sprintf(" @%s at %s ", authorLogin, createdAt)

	fmt.Printf("  <!--%s%s--->\n\n",
		strings.Repeat("-", 69 - len(attribution)),
		attribution)
	value = value["comments"].(map[string]interface{})
	comments := value["nodes"].([]interface{})
	wrapper := textwrap.NewTextWrap().SetWidth(79)

	for _, commenti := range comments {
		comment := commenti.(map[string]interface{})
		body := comment["body"].(string)
		author := comment["author"].(map[string]interface{})
		authorLogin := author["login"].(string)
		createdAt := comment["createdAt"].(string)

		fmt.Printf("%s\n", wrapper.Fill(body))
		attribution := fmt.Sprintf(" @%s at %s ", authorLogin,
			createdAt)
		fmt.Printf("  <!--%s%s--->\n\n",
			strings.Repeat("-", 69 - len(attribution)),
			attribution)
	}
}
