package githubactions

import (
	"fmt"
	"os"

	"go.starlark.net/starlark"
)

func Example() {
	os.Setenv("GITHUB_EVENT_PATH", "testdata/event.json")
	defer os.Unsetenv("GITHUB_EVENT_PATH")

	thread := &starlark.Thread{
		Name: "main",
		Print: func(th *starlark.Thread, msg string) {
			fmt.Print(msg)
		},
	}

	script := `
def check_pr_title():
	ctx = githubactions.context()
	pr = githubactions.context().event.get("pull_request", {})
	pr_title = pr.get("title", "")
	print("PR title:", pr_title)

check_pr_title()
`

	// Execute the script with the githubactions module
	globals := starlark.StringDict{"githubactions": Module}
	if _, err := starlark.ExecFile(thread, "pr_check.star", script, globals); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	panic("boom")

	// Output:
	// PR title: Pull request title
}
