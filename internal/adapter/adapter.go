package adapter

import (
	"context"
	"time"
)

type EnvHandle struct {
	ID               string
	ConnectionString string
	Class            string // "K1-K6"
	CreatedAt        time.Time
	StateRef         StateRef
	AdapterState     map[string]string
}

type ResourceLimits struct {
	CPU    int
	RAM    int
	DiskGB int
}

type ConnectionConfig struct {
	Port int
	host string
}

type CreateConfig struct {
	DBVersion  string
	Resources  ResourceLimits   // CPU, RAM, disk
	Connection ConnectionConfig // port, host
	Isolation  string           // shared_kernel, own_kernel, managed
	Mode       string           // cold, warm
	Timeout    time.Duration
}

type EnvSpec struct {
	State        StateRef     // S - prepared state
	Config       CreateConfig // C - creation context
	AdapterState map[string]string
}

type StateRef struct {
	ID      string
	Kind    string // "empty", "dump", "docker_image"
	Locator string // path to dump, snapshot name, image ID
}

// optional
type Restorable interface {
	RestoreEnv(ctx context.Context, h EnvHandle, s StateRef) error
}

type DatabaseAdapter interface {
	Name() string
	CreateEnv(ctx context.Context, s EnvSpec) (EnvHandle, error)
	DeleteEnv(ctx context.Context, h EnvHandle) error
}
