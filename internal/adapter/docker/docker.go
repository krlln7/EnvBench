package docker

import (
	"context"
	"fmt"
	"net/netip"
	"time"

	"envbench/internal/adapter"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
)

type Adapter struct {
	cli *client.Client
}

func init() {
	adapter.Register("docker", func() (adapter.DatabaseAdapter, error) {
		return New()
	})
}

func New() (*Adapter, error) {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		return nil, err
	}
	return &Adapter{cli: cli}, nil
}

func (a *Adapter) Name() string  { return "docker-postgres" }
func (a *Adapter) Class() string { return "K1" }

func (a *Adapter) CreateEnv(ctx context.Context, spec adapter.EnvSpec) (adapter.EnvHandle, error) {
	if spec.State.ID == "" {
		return adapter.EnvHandle{}, fmt.Errorf("StateRef.LogicalID is required")
	}

	pgPort := network.MustParsePort("5432/tcp")

	hostPortStr := ""
	if spec.Config.Connection.Port != 0 {
		hostPortStr = fmt.Sprintf("%d", spec.Config.Connection.Port)
	}

	resp, err := a.cli.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &container.Config{
			Image: "postgres:" + spec.Config.DBVersion,
			Env: []string{
				"POSTGRES_PASSWORD=bench",
				"POSTGRES_DB=bench",
			},
			ExposedPorts: network.PortSet{pgPort: struct{}{}},
		},
		HostConfig: &container.HostConfig{
			PortBindings: network.PortMap{
				pgPort: []network.PortBinding{{
					HostIP:   netip.MustParseAddr("127.0.0.1"),
					HostPort: hostPortStr,
				}},
			},
			NanoCPUs: int64(spec.Config.Resources.CPU) * 1_000_000_000,
			// other specs?
		},
	})
	if err != nil {
		return adapter.EnvHandle{}, fmt.Errorf("container create: %w", err)
	}

	if _, err := a.cli.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{}); err != nil {
		_, _ = a.cli.ContainerRemove(ctx, resp.ID, client.ContainerRemoveOptions{Force: true})
		return adapter.EnvHandle{}, fmt.Errorf("container start: %w", err)
	}

	inspect, err := a.cli.ContainerInspect(ctx, resp.ID, client.ContainerInspectOptions{})
	if err != nil {
		return adapter.EnvHandle{}, fmt.Errorf("inspect: %w", err)
	}
	actualPort := inspect.Container.NetworkSettings.Ports[pgPort][0].HostPort
	dsn := fmt.Sprintf("postgres://postgres:bench@127.0.0.1:%s/bench?sslmode=disable", actualPort)

	return adapter.EnvHandle{
		ID:               resp.ID,
		ConnectionString: dsn,
		Class:            "K1",
		CreatedAt:        time.Now(),
		StateRef:         spec.State,
		AdapterState:     map[string]string{"container_id": resp.ID},
	}, nil
}

func (a *Adapter) DeleteEnv(ctx context.Context, h adapter.EnvHandle) error {
	_, _ = a.cli.ContainerRemove(ctx, h.ID, client.ContainerRemoveOptions{Force: true, RemoveVolumes: true})
	return nil
}
