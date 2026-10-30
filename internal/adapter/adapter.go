package adapter

import (
	"context"
	"database/sql"
	"time"
)

type EnvHandle struct {
	ID               string
	ConnectionString string
	Class            string // "K1-K6"
	CreatedAt        time.Time
	StateRef         StateRef
}

type CreateConfig struct {
	DBVersion string
	// Resources   ResourceLimits    // CPU, RAM, disk
	// Connection  ConnectionConfig  // port, host
	// Isolation   IsolationPolicy   // shared_kernel / own_kernel / managed
	Timeout time.Duration
}

type EnvSpec struct {
	State  StateRef     // S - prepared state
	Config CreateConfig // C - creation context
}

type StateRef struct {
	Kind     string            // "empty", "dump", "docker_image"
	Locator  string            // path to dump, snapshot name, image ID
	Metadata map[string]string // size, hash, scheme version
}

type DatabaseAdapter interface {
	Name() string
	CreateEnv(ctx context.Context, s EnvSpec) (EnvHandle, error)
	Connect(ctx context.Context, h EnvHandle) (*sql.Conn, error)
	DeleteEnv(ctx context.Context, h EnvHandle) error
}
