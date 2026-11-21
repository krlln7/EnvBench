package scenario

import (
	"context"
	"fmt"

	"envbench/internal/adapter"
	"envbench/internal/engine"
	"envbench/internal/result"

	"gopkg.in/yaml.v3"
)

type S4 struct {
	MutationSQL string `yaml:"mutation_sql"`
}

type S4Result struct {
	IsolationOK bool
	ReferenceOK bool
	E1Report    string
	E2Report    string
	RefReport   string
}

type s4Params struct {
	MutationSQL string `yaml:"mutation_sql"`
}

func init() {
	Register("s4", runS4)
}

// 1) Create E1, E2
// 2) Check isolation property (equality to etalon)
// 3) Mutate E1
// 4) Check E2 (not changed)
// 5) Clear
func runS4(ctx context.Context, r *engine.Engine, spec adapter.EnvSpec, params yaml.Node) (result.Result, error) {
	var p s4Params
	if err := params.Decode(&p); err != nil {
		return result.Result{}, fmt.Errorf("s4 params: %w", err)
	}
	if p.MutationSQL == "" {
		return result.Result{}, fmt.Errorf("s4: mutation_sql is required")
	}

	res := result.Result{Checks: map[string]string{}}

	eRef, _, err := r.Create(ctx, spec)
	if err != nil {
		return res, fmt.Errorf("create ref: %w", err)
	}
	defer r.Delete(ctx, eRef)

	e1, _, err := r.Create(ctx, spec)
	if err != nil {
		return res, fmt.Errorf("create e1: %w", err)
	}
	defer r.Delete(ctx, e1)

	e2, _, err := r.Create(ctx, spec)
	if err != nil {
		return res, fmt.Errorf("create e2: %w", err)
	}
	defer r.Delete(ctx, e2)

	rep1, err := r.Check(ctx, e1, eRef.ConnectionString)
	if err != nil {
		return res, err
	}
	rep2, err := r.Check(ctx, e2, eRef.ConnectionString)
	if err != nil {
		return res, err
	}
	res.Checks["E1_pre"] = rep1.Verdict
	res.Checks["E2_pre"] = rep2.Verdict

	if err := r.ExecSQL(ctx, e1.ConnectionString, p.MutationSQL); err != nil {
		return res, fmt.Errorf("mutation: %w", err)
	}

	rep2After, err := r.Check(ctx, e2, eRef.ConnectionString)
	if err != nil {
		return res, err
	}
	res.Checks["E2_post"] = rep2After.Verdict

	// check etalon
	rep1After, err := r.Check(ctx, e1, eRef.ConnectionString)
	if err != nil {
		return res, err
	}
	res.Checks["E1_post"] = rep1After.Verdict

	if rep2After.Verdict == "equivalent" {
		res.Verdict = "admitted"
	} else {
		res.Verdict = "correctness_failure"
	}
	return res, nil
}
