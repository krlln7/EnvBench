package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

func WaitReady(ctx context.Context, dsn string, timeout time.Duration) (time.Time, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return time.Time{}, fmt.Errorf("ready timeout: %w", ctx.Err())
		case <-ticker.C:
			conn, err := pgx.Connect(ctx, dsn)
			if err != nil {
				continue
			}
			var one int
			err = conn.QueryRow(ctx, "SELECT 1").Scan(&one)
			_ = conn.Close(ctx)
			if err == nil && one == 1 {
				return time.Now(), nil
			}
		}
	}
}
