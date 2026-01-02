package githubactions_test

import (
	"fmt"
	"log"
	"os"

	githubactions "github.com/AlekSi/starlark-githubactions"
	gogithubactions "github.com/sethvargo/go-githubactions"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

func Example() {
	// Create a Starlark module for this example.
	getenv := func(key string) string {
		switch key {
		case "GITHUB_EVENT_PATH":
			return "testdata/event.json"
		default:
			return ""
		}
	}
	module := githubactions.NewModule(
		"githubactions",
		githubactions.New(gogithubactions.New(
			gogithubactions.WithWriter(os.Stdout),
			gogithubactions.WithGetenv(getenv),
		)),
	)

	// Add module to the predeclared global environment.
	// Most users should use githubactions.Module instead.
	predeclared := starlark.StringDict{
		"githubactions": module,
	}

	script := `
def merge_method():
	ctx = githubactions.context()
	pr = ctx.event.get("pull_request", {})
	return pr.get("auto_merge", {}).get("merge_method", "")

print(merge_method())
`

	opts := &syntax.FileOptions{}
	thread := &starlark.Thread{
		Print: func(th *starlark.Thread, msg string) {
			fmt.Print(msg)
		},
	}
	if _, err := starlark.ExecFileOptions(opts, thread, "", script, predeclared); err != nil {
		log.Fatal(err)
	}

	// Output:
	// squash
}
