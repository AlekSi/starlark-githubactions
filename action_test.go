package githubactions

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/AlekSi/should"
	"github.com/AlekSi/should/must"
	"github.com/sethvargo/go-githubactions"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// setup prepares a Starlark thread and GitHub Actions module for testing.
func setup(tb testing.TB, w io.Writer, getenv githubactions.GetenvFunc) (*starlark.Thread, *starlarkstruct.Module, githubactions.GetenvFunc) {
	tb.Helper()

	must.NotBeZerof(tb, w, "writer must not be nil")

	files := map[string]string{
		"GITHUB_ENV":          "",
		"GITHUB_OUTPUT":       "",
		"GITHUB_PATH":         "",
		"GITHUB_STATE":        "",
		"GITHUB_STEP_SUMMARY": "",
	}
	for e := range files {
		f, err := os.CreateTemp(tb.TempDir(), strings.ToLower(e))
		must.BeZero(tb, err)
		f.Close()
		files[e] = f.Name()
	}

	newGetenv := func(key string) string {
		if getenv != nil {
			if v := getenv(key); v != "" {
				return v
			}
		}

		fn := files[key]
		must.NotBeZero(tb, fn)
		return fn
	}

	th := &starlark.Thread{
		Name: tb.Name(),
	}

	m := module(tb.Name(), newAction(w, newGetenv))
	return th, m, newGetenv
}

func TestLog(t *testing.T) {
	var buf bytes.Buffer
	th, m, _ := setup(t, &buf, nil)

	res, err := starlark.Call(th, m.Members["log"], starlark.Tuple{starlark.String("log message")}, nil)
	must.BeZero(t, err)
	should.BeEqual(t, res, starlark.None)
	should.BeEqual(t, buf.String(), "log message\n")
}

func TestDebug(t *testing.T) {
	var buf bytes.Buffer
	th, m, _ := setup(t, &buf, nil)

	res, err := starlark.Call(th, m.Members["debug"], starlark.Tuple{starlark.String("debug message")}, nil)
	must.BeZero(t, err)
	should.BeEqual(t, res, starlark.None)
	should.BeEqual(t, buf.String(), "::debug::debug message\n")
}

func TestNotice(t *testing.T) {
	var buf bytes.Buffer
	th, m, _ := setup(t, &buf, nil)

	res, err := starlark.Call(th, m.Members["notice"], starlark.Tuple{starlark.String("notice message")}, nil)
	must.BeZero(t, err)
	should.BeEqual(t, res, starlark.None)
	should.BeEqual(t, buf.String(), "::notice::notice message\n")
}

func TestWarning(t *testing.T) {
	var buf bytes.Buffer
	th, m, _ := setup(t, &buf, nil)

	res, err := starlark.Call(th, m.Members["warning"], starlark.Tuple{starlark.String("warning message")}, nil)
	must.BeZero(t, err)
	should.BeEqual(t, res, starlark.None)
	should.BeEqual(t, buf.String(), "::warning::warning message\n")
}

func TestError(t *testing.T) {
	var buf bytes.Buffer
	th, m, _ := setup(t, &buf, nil)

	res, err := starlark.Call(th, m.Members["error"], starlark.Tuple{starlark.String("error message")}, nil)
	must.BeZero(t, err)
	should.BeEqual(t, res, starlark.None)
	should.BeEqual(t, buf.String(), "::error::error message\n")
}

func TestAddMatcher(t *testing.T) {
	var buf bytes.Buffer
	th, m, _ := setup(t, &buf, nil)

	res, err := starlark.Call(th, m.Members["add_matcher"], starlark.Tuple{starlark.String("matcher-path")}, nil)
	must.BeZero(t, err)
	should.BeEqual(t, res, starlark.None)
	should.BeEqual(t, buf.String(), "::add-matcher::matcher-path\n")
}

func TestRemoveMatcher(t *testing.T) {
	var buf bytes.Buffer
	th, m, _ := setup(t, &buf, nil)

	res, err := starlark.Call(th, m.Members["remove_matcher"], starlark.Tuple{starlark.String("matcher-owner")}, nil)
	must.BeZero(t, err)
	should.BeEqual(t, res, starlark.None)
	should.BeEqual(t, buf.String(), "::remove-matcher owner=matcher-owner::\n")
}

func TestAddMask(t *testing.T) {
	var buf bytes.Buffer
	th, m, _ := setup(t, &buf, nil)

	res, err := starlark.Call(th, m.Members["add_mask"], starlark.Tuple{starlark.String("secret-value")}, nil)
	must.BeZero(t, err)
	should.BeEqual(t, res, starlark.None)
	should.BeEqual(t, buf.String(), "::add-mask::secret-value\n")
}

func TestAddStepSummary(t *testing.T) {
	var buf bytes.Buffer
	th, m, getenv := setup(t, &buf, nil)

	res, err := starlark.Call(th, m.Members["add_step_summary"], starlark.Tuple{starlark.String("summary message")}, nil)
	must.BeZero(t, err)
	should.BeEqual(t, res, starlark.None)
	should.BeEqual(t, buf.String(), "")

	b, err := os.ReadFile(getenv("GITHUB_STEP_SUMMARY"))
	must.BeZero(t, err)
	should.BeEqual(t, string(b), "summary message\n")
}

func TestGroup(t *testing.T) {
	var buf bytes.Buffer
	th, m, _ := setup(t, &buf, nil)

	res, err := starlark.Call(th, m.Members["group"], starlark.Tuple{starlark.String("group title")}, nil)
	must.BeZero(t, err)
	should.BeEqual(t, res, starlark.None)
	should.BeEqual(t, buf.String(), "::group::group title\n")
}

func TestEndGroup(t *testing.T) {
	var buf bytes.Buffer
	th, m, _ := setup(t, &buf, nil)

	res, err := starlark.Call(th, m.Members["end_group"], starlark.Tuple{}, nil)
	must.BeZero(t, err)
	should.BeEqual(t, res, starlark.None)
	should.BeEqual(t, buf.String(), "::endgroup::\n")
}

func TestGetInput(t *testing.T) {
	var buf bytes.Buffer
	th, m, _ := setup(t, &buf, func(key string) string {
		must.BeEqual(t, key, "INPUT_MY_INPUT")
		return "test value"
	})

	res, err := starlark.Call(th, m.Members["get_input"], starlark.Tuple{starlark.String("my_input")}, nil)
	must.BeZero(t, err)
	should.BeEqual(t, res, starlark.String("test value"))
	should.BeEqual(t, buf.String(), "")
}

func TestSetOutput(t *testing.T) {
	var buf bytes.Buffer
	th, m, getenv := setup(t, &buf, nil)

	res, err := starlark.Call(th, m.Members["set_output"], starlark.Tuple{starlark.String("my_output"), starlark.String("output value")}, nil)
	must.BeZero(t, err)
	should.BeEqual(t, res, starlark.None)
	should.BeEqual(t, buf.String(), "")

	b, err := os.ReadFile(getenv("GITHUB_OUTPUT"))
	must.BeZero(t, err)
	should.BeEqual(t, string(b), "my_output<<_GitHubActionsFileCommandDelimeter_\noutput value\n_GitHubActionsFileCommandDelimeter_\n")
}

func TestSaveState(t *testing.T) {
	var buf bytes.Buffer
	th, m, getenv := setup(t, &buf, nil)

	res, err := starlark.Call(th, m.Members["save_state"], starlark.Tuple{starlark.String("my_state"), starlark.String("state value")}, nil)
	must.BeZero(t, err)
	should.BeEqual(t, res, starlark.None)
	should.BeEqual(t, buf.String(), "")

	b, err := os.ReadFile(getenv("GITHUB_STATE"))
	must.BeZero(t, err)
	should.BeEqual(t, string(b), "my_state<<_GitHubActionsFileCommandDelimeter_\nstate value\n_GitHubActionsFileCommandDelimeter_\n")
}

func TestSetEnv(t *testing.T) {
	var buf bytes.Buffer
	th, m, getenv := setup(t, &buf, nil)

	res, err := starlark.Call(th, m.Members["set_env"], starlark.Tuple{starlark.String("MY_VAR"), starlark.String("my value")}, nil)
	must.BeZero(t, err)
	should.BeEqual(t, res, starlark.None)
	should.BeEqual(t, buf.String(), "")

	b, err := os.ReadFile(getenv("GITHUB_ENV"))
	must.BeZero(t, err)
	should.BeEqual(t, string(b), "MY_VAR<<_GitHubActionsFileCommandDelimeter_\nmy value\n_GitHubActionsFileCommandDelimeter_\n")
}

func TestAddPath(t *testing.T) {
	var buf bytes.Buffer
	th, m, getenv := setup(t, &buf, nil)

	res, err := starlark.Call(th, m.Members["add_path"], starlark.Tuple{starlark.String("/new/path")}, nil)
	must.BeZero(t, err)
	should.BeEqual(t, res, starlark.None)
	should.BeEqual(t, buf.String(), "")

	b, err := os.ReadFile(getenv("GITHUB_PATH"))
	must.BeZero(t, err)
	should.BeEqual(t, string(b), "/new/path\n")
}

func TestContext(t *testing.T) {
	getenv := func(key string) string {
		return map[string]string{
			// no GITHUB_ENV, GITHUB_PATH, GITHUB_STEP_SUMMARY
			"GITHUB_ACTION":            "test-action",
			"GITHUB_ACTION_PATH":       "/path/to/action",
			"GITHUB_ACTION_REPOSITORY": "owner/repo",
			"GITHUB_ACTIONS":           "true",
			"GITHUB_ACTOR":             "testactor",
			"GITHUB_ACTOR_ID":          "12345",
			"GITHUB_API_URL":           "https://api.github.com",
			"GITHUB_BASE_REF":          "main",
			"GITHUB_EVENT_NAME":        "push",
			"GITHUB_EVENT_PATH":        "testdata/event.json",
			"GITHUB_GRAPHQL_URL":       "https://api.github.com/graphql",
			"GITHUB_HEAD_REF":          "feature",
			"GITHUB_JOB":               "test-job",
			"GITHUB_REF":               "refs/heads/main",
			"GITHUB_REF_NAME":          "main",
			"GITHUB_REF_PROTECTED":     "false",
			"GITHUB_REF_TYPE":          "branch",
			"GITHUB_REPOSITORY":        "owner/repo",
			"GITHUB_REPOSITORY_OWNER":  "owner",
			"GITHUB_RETENTION_DAYS":    "30",
			"GITHUB_RUN_ATTEMPT":       "1",
			"GITHUB_RUN_ID":            "123456",
			"GITHUB_RUN_NUMBER":        "42",
			"GITHUB_SERVER_URL":        "https://github.com",
			"GITHUB_SHA":               "abc123",
			"GITHUB_TRIGGERING_ACTOR":  "testactor",
			"GITHUB_WORKFLOW":          "CI",
			"GITHUB_WORKSPACE":         "/workspace",
		}[key]
	}

	var buf bytes.Buffer
	th, m, _ := setup(t, &buf, getenv)

	res, err := starlark.Call(th, m.Members["context"], starlark.Tuple{}, nil)
	must.BeZero(t, err)
	should.BeEqual(t, buf.String(), "")

	s, ok := res.(*starlarkstruct.Struct)
	must.NotBeZero(t, ok)

	v, err := s.Attr("actions")
	must.BeZero(t, err)
	should.BeEqual(t, v, starlark.True)

	v, err = s.Attr("run_number")
	must.BeZero(t, err)
	should.BeEqual(t, v, starlark.MakeInt(42))

	v, err = s.Attr("event")
	must.BeZero(t, err)

	event, ok := v.(*starlark.Dict)
	must.NotBeZero(t, ok)

	_ = event
}
