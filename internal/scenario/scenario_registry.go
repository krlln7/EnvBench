package scenario

import (
	"context"
	"envbench/internal/adapter"
	"envbench/internal/engine"
	"envbench/internal/result"
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

type RunnerFunc func(ctx context.Context, e *engine.Engine, spec adapter.EnvSpec, params yaml.Node) (result.Result, error)

var registry = map[string]RunnerFunc{}

func Register(kind string, fn RunnerFunc) {
	if _, ok := registry[kind]; ok {
		panic("scenario already registered: " + kind)
	}
	registry[kind] = fn
}

func Run(ctx context.Context, e *engine.Engine, sc Scenario) (result.Result, error) {
	fn, ok := registry[sc.Run.Kind]
	if !ok {
		return result.Result{}, fmt.Errorf("unknown scenario kind %q (registered: %v)", sc.Run.Kind, Names())
	}
	return fn(ctx, e, sc.ToEnvSpec(), sc.Run.Params)
}

func Names() []string {
	out := make([]string, 0, len(registry))
	for k := range registry {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
