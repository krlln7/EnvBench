package engine

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5"
)

func pgxConnect(ctx context.Context, dsn string) (*pgx.Conn, error) {
	return pgx.Connect(ctx, dsn)
}

func execFile(ctx context.Context, dsn, path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	_, err = conn.Exec(ctx, string(b))
	return err
}
