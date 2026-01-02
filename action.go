package githubactions

import (
	"encoding/json"
	"maps"
	"math/big"
	"os"
	"slices"

	"github.com/AlekSi/lazyerrors"
	"github.com/sethvargo/go-githubactions"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// Action wraps [githubactions.Action] for Starlark.
type Action struct {
	a *githubactions.Action
}

// New creates a new [Action].
func New(a *githubactions.Action) *Action {
	return &Action{a: a}
}

// log logs a message using fmt.Printf-like function.
func (a *Action) log(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple, logf func(msg string, args ...any)) (string, error) {
	var msg string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "msg", &msg); err != nil {
		return msg, err
	}

	logf("%s", msg)
	return msg, nil
}

// Log prints a message without level annotation.
//
// The caller is expected to format the message using string interpolation with % operator,
// string.format method, or other means.
func (a *Action) Log(th *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	_, err := a.log(th, fn, args, kwargs, a.a.Infof)
	return starlark.None, err
}

// Debug prints a debug-level message.
// See https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#setting-a-debug-message.
//
// The caller is expected to format the message using string interpolation with % operator,
// string.format method, or other means.
func (a *Action) Debug(th *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	_, err := a.log(th, fn, args, kwargs, a.a.Debugf)
	return starlark.None, err
}

// Notice prints a notice-level message.
// See https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#setting-a-notice-message.
//
// The caller is expected to format the message using string interpolation with % operator,
// string.format method, or other means.
func (a *Action) Notice(th *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	_, err := a.log(th, fn, args, kwargs, a.a.Noticef)
	return starlark.None, err
}

// Warning prints a warning-level message.
// See https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#setting-a-warning-message.
//
// The caller is expected to format the message using string interpolation with % operator,
// string.format method, or other means.
func (a *Action) Warning(th *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	_, err := a.log(th, fn, args, kwargs, a.a.Warningf)
	return starlark.None, err
}

// Error prints a error-level message.
// See https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#setting-an-error-message.
//
// The caller is expected to format the message using string interpolation with % operator,
// string.format method, or other means.
func (a *Action) Error(th *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	_, err := a.log(th, fn, args, kwargs, a.a.Errorf)
	return starlark.None, err
}

// Fatal prints a message using [action.Error] and fails the Starlark thread.
//
// The caller is expected to format the message using string interpolation with % operator,
// string.format method, or other means.
func (a *Action) Fatal(th *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	msg, err := a.log(th, fn, args, kwargs, a.a.Errorf) // not Fatalf

	if err == nil {
		th.Cancel(msg)
	}

	return starlark.None, err
}

// AddMatcher adds a new matcher with the given file path.
func (a *Action) AddMatcher(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var path string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "path", &path); err != nil {
		return nil, err
	}

	a.a.AddMatcher(path)
	return starlark.None, nil
}

// RemoveMatcher removes a matcher with the given owner name.
func (a *Action) RemoveMatcher(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var owner string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "owner", &owner); err != nil {
		return nil, err
	}

	a.a.RemoveMatcher(owner)
	return starlark.None, nil
}

// AddMask adds a new field mask for the given value.
// After called, future attempts to log the value will be replaced with "***" in log output.
// See https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#masking-a-value-in-a-log.
func (a *Action) AddMask(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var value string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "value", &value); err != nil {
		return nil, err
	}

	a.a.AddMask(value)
	return starlark.None, nil
}

// AddStepSummary writes the given markdown to the job summary.
// If a job summary already exists, this value is appended.
// See https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#adding-a-job-summary.
func (a *Action) AddStepSummary(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var summary string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "summary", &summary); err != nil {
		return nil, err
	}

	a.a.AddStepSummary(summary)
	return starlark.None, nil
}

// Group starts a new collapsable region up to the next endgroup invocation.
// See https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#grouping-log-lines.
func (a *Action) Group(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var title string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "title", &title); err != nil {
		return nil, err
	}

	a.a.Group(title)
	return starlark.None, nil
}

// EndGroup ends the current group.
// See https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#grouping-log-lines.
func (a *Action) EndGroup(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs); err != nil {
		return nil, err
	}

	a.a.EndGroup()
	return starlark.None, nil
}

// GetInput gets the input by the given name.
// Returns the empty string if the input is not defined.
func (a *Action) GetInput(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var name string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "name", &name); err != nil {
		return nil, err
	}

	return starlark.String(a.a.GetInput(name)), nil
}

// SetOutput sets an output parameter.
// See https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#setting-an-output-parameter.
func (a *Action) SetOutput(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var name, value string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "name", &name, "value", &value); err != nil {
		return nil, err
	}

	a.a.SetOutput(name, value)
	return starlark.None, nil
}

// SaveState saves state to be used in the "finally" post job entry point.
// See https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#sending-values-to-the-pre-and-post-actions.
func (a *Action) SaveState(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var name, value string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "name", &name, "value", &value); err != nil {
		return nil, err
	}

	a.a.SaveState(name, value)
	return starlark.None, nil
}

// SetEnv sets an environment variable.
// See https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#setting-an-environment-variable.
func (a *Action) SetEnv(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var name, value string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "name", &name, "value", &value); err != nil {
		return nil, err
	}

	a.a.SetEnv(name, value)
	return starlark.None, nil
}

// AddPath adds a new system path to the PATH environment variable.
// See https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#adding-a-system-path.
func (a *Action) AddPath(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var path string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "path", &path); err != nil {
		return nil, err
	}

	a.a.AddPath(path)
	return starlark.None, nil
}

// Context returns the GitHub Actions Context as a Starlark struct.
func (a *Action) Context(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs); err != nil {
		return nil, err
	}

	ctx, err := a.a.Context()
	if err != nil {
		return nil, err
	}

	// do not use ctx.Event; decode ourselves to convert to Starlark values
	event, err := readEvent(ctx.EventPath)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	res := starlarkstruct.FromStringDict(starlark.String("context"), starlark.StringDict{
		"action":            starlark.String(ctx.Action),
		"action_path":       starlark.String(ctx.ActionPath),
		"action_repository": starlark.String(ctx.ActionRepository),
		"actions":           starlark.Bool(ctx.Actions),
		"actor":             starlark.String(ctx.Actor),
		"actor_id":          starlark.String(ctx.ActorID),
		"api_url":           starlark.String(ctx.APIURL),
		"base_ref":          starlark.String(ctx.BaseRef),
		"env":               starlark.String(ctx.Env),
		"event":             event,
		"event_name":        starlark.String(ctx.EventName),
		"event_path":        starlark.String(ctx.EventPath),
		"graphql_url":       starlark.String(ctx.GraphqlURL),
		"head_ref":          starlark.String(ctx.HeadRef),
		"job":               starlark.String(ctx.Job),
		"path":              starlark.String(ctx.Path),
		"ref":               starlark.String(ctx.Ref),
		"ref_name":          starlark.String(ctx.RefName),
		"ref_protected":     starlark.Bool(ctx.RefProtected),
		"ref_type":          starlark.String(ctx.RefType),
		"repository":        starlark.String(ctx.Repository),
		"repository_owner":  starlark.String(ctx.RepositoryOwner),
		"retention_days":    starlark.MakeInt64(ctx.RetentionDays),
		"run_attempt":       starlark.MakeInt64(ctx.RunAttempt),
		"run_id":            starlark.MakeInt64(ctx.RunID),
		"run_number":        starlark.MakeInt64(ctx.RunNumber),
		"server_url":        starlark.String(ctx.ServerURL),
		"sha":               starlark.String(ctx.SHA),
		"step_summary":      starlark.String(ctx.StepSummary),
		"triggering_actor":  starlark.String(ctx.TriggeringActor),
		"workflow":          starlark.String(ctx.Workflow),
		"workspace":         starlark.String(ctx.Workspace),
	})

	res.Freeze()
	return res, nil
}

// readEvent reads and decodes the GitHub event JSON file at the given path.
func readEvent(path string) (starlark.Value, error) {
	if path == "" {
		return starlark.None, nil
	}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return starlark.None, nil
		}

		return nil, lazyerrors.Error(err)
	}

	defer f.Close()

	d := json.NewDecoder(f)
	d.UseNumber()

	var e map[string]any
	if err = d.Decode(&e); err != nil {
		return nil, lazyerrors.Error(err)
	}

	event, err := jsonToStarlark(e)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return event, nil
}

// jsonToStarlark converts a Go value that could be decoded from JSON
// ([]any, map[string]any, nil/null, bool, string, [json.Number])
// to a Starlark value.
func jsonToStarlark(v any) (starlark.Value, error) {
	switch v := v.(type) {
	case []any:
		elems := make([]starlark.Value, len(v))
		for i, e := range v {
			sv, err := jsonToStarlark(e)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			elems[i] = sv
		}

		list := starlark.NewList(elems)
		return list, nil

	case map[string]any:
		dict := starlark.NewDict(len(v))
		for _, k := range slices.Sorted(maps.Keys(v)) {
			sv, err := jsonToStarlark(v[k])
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			if err = dict.SetKey(starlark.String(k), sv); err != nil {
				return nil, lazyerrors.Error(err)
			}
		}

		return dict, nil

	case nil:
		return starlark.None, nil

	case bool:
		return starlark.Bool(v), nil

	case string:
		return starlark.String(v), nil

	case json.Number:
		if i, ok := new(big.Int).SetString(string(v), 10); ok {
			return starlark.MakeBigInt(i), nil
		}

		f, err := v.Float64()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return starlark.Float(f), nil

	default:
		return nil, lazyerrors.Errorf("unsupported JSON value type: %T", v)
	}
}
