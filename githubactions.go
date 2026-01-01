// Package githubactions provides [starlark-go] wrappers for [go-githubactions].
//
// [starlark-go]: https://github.com/google/starlark-go
// [go-githubactions]: https://github.com/sethvargo/go-githubactions
package githubactions

import (
	"os"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// module constructs a Starlark module for the given [action].
func module(name string, a *action) *starlarkstruct.Module {
	m := &starlarkstruct.Module{
		Name:    name,
		Members: make(starlark.StringDict),
	}

	for _, b := range []*starlark.Builtin{
		starlark.NewBuiltin("log", a.Log),
		starlark.NewBuiltin("debug", a.Debug),
		starlark.NewBuiltin("notice", a.Notice),
		starlark.NewBuiltin("warning", a.Warning),
		starlark.NewBuiltin("error", a.Error),
		starlark.NewBuiltin("fatal", a.Fatal),

		starlark.NewBuiltin("add_matcher", a.AddMatcher),
		starlark.NewBuiltin("remove_matcher", a.RemoveMatcher),

		starlark.NewBuiltin("add_mask", a.AddMask),

		starlark.NewBuiltin("add_step_summary", a.AddStepSummary),

		starlark.NewBuiltin("group", a.Group),
		starlark.NewBuiltin("end_group", a.EndGroup),

		starlark.NewBuiltin("get_input", a.GetInput),
		starlark.NewBuiltin("set_output", a.SetOutput),

		starlark.NewBuiltin("save_state", a.SaveState),

		starlark.NewBuiltin("set_env", a.SetEnv),
		starlark.NewBuiltin("add_path", a.AddPath),

		starlark.NewBuiltin("context", a.Context),
	} {
		m.Members[b.Name()] = b
	}

	return m
}

// Module is the GitHub Actions Starlark module.
var Module = module("githubactions", newAction(os.Stdout, os.Getenv))
