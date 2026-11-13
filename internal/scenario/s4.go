package scenario

import (
	"context"
	"fmt"

	"envbench/internal/adapter"
	"envbench/internal/runner"
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

// 1) Create E1, E2
// 2) Check isolation property (equality to etalon)
// 3) Mutate E1
// 4) Check E2 (not changed)
// 5) Clear
func RunS4(ctx context.Context, r *runner.Runner, spec adapter.EnvSpec,
	referenceDSN string, mutationSQL string) (S4Result, error) {

	var res S4Result

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

	rep1, err := r.Check(ctx, e1, referenceDSN)
	if err != nil {
		return res, fmt.Errorf("check e1: %w", err)
	}
	rep2, err := r.Check(ctx, e2, referenceDSN)
	if err != nil {
		return res, fmt.Errorf("check e2: %w", err)
	}
	res.E1Report = rep1.Verdict
	res.E2Report = rep2.Verdict

	if err := r.ExecSQL(ctx, e1.ConnectionString, mutationSQL); err != nil {
		return res, fmt.Errorf("mutation: %w", err)
	}

	rep2After, err := r.Check(ctx, e2, referenceDSN)
	if err != nil {
		return res, fmt.Errorf("check e2 after: %w", err)
	}
	res.IsolationOK = rep2After.Verdict == "equivalent"

	// check etalon
	repRef, err := r.Check(ctx, e1, referenceDSN)
	if err != nil {
		return res, fmt.Errorf("check ref: %w", err)
	}
	res.RefReport = repRef.Verdict

	return res, nil
}
