package engine

import (
	"context"
	"fmt"
	"time"

	"envbench/internal/adapter"
	"envbench/internal/checker"
)

type Engine struct {
	Adapter adapter.DatabaseAdapter
}

func (e *Engine) Create(ctx context.Context, spec adapter.EnvSpec) (adapter.EnvHandle, time.Duration, error) {
	t0 := time.Now()
	h, err := e.Adapter.CreateEnv(ctx, spec)
	if err != nil {
		return adapter.EnvHandle{}, 0, err
	}
	t1, err := WaitReady(ctx, h.ConnectionString, spec.Config.Timeout)
	if err != nil {
		_ = e.Adapter.DeleteEnv(context.Background(), h)
		return adapter.EnvHandle{}, 0, err
	}
	return h, t1.Sub(t0), nil
}

func (e *Engine) Delete(ctx context.Context, h adapter.EnvHandle) error {
	cctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return e.Adapter.DeleteEnv(cctx, h)
}

func (e *Engine) Check(ctx context.Context, h adapter.EnvHandle, referenceDSN string) (checker.Report, error) {
	return checker.EqEnv(ctx, h.ConnectionString, referenceDSN)
}

func (e *Engine) ExecSQL(ctx context.Context, dsn, sqlText string) error {
	conn, err := pgxConnect(ctx, dsn)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	_, err = conn.Exec(ctx, sqlText)
	return err
}

func (e *Engine) SetupState(ctx context.Context, dsn string, s adapter.StateRef) error {
	if s.Kind == "empty" {
		return nil
	}
	if s.Kind == "dump" {
		return execFile(ctx, dsn, s.Locator)
	}
	return fmt.Errorf("unsupported state kind: %s", s.Kind)
}
