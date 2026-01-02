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
			return "event.json"
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
def check_pr():
	ctx = githubactions.context()
	pr = ctx.event.get("pull_request", {})
	if not pr:
		fail("Not a 'pull_request' event")

	merge_method = pr.get("auto_merge", {}).get("merge_method", "")
	if not merge_method:
		fail("Auto-merge should be enabled")

	print("Merge method:", merge_method)

check_pr()
`

	opts := &syntax.FileOptions{}
	thread := &starlark.Thread{
		Print: func(th *starlark.Thread, msg string) {
			fmt.Println(msg)
		},
	}

	if _, err := starlark.ExecFileOptions(opts, thread, "", script, predeclared); err != nil {
		log.Fatal(err)
	}
	// Output:
	// Merge method: squash
}
