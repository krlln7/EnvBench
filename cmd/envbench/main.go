package main

import (
	"context"
	"encoding/json"
	"envbench/internal/adapter"
	_ "envbench/internal/adapter/docker" // for init()
	"envbench/internal/runner"
	"envbench/internal/scenario"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: envbench <scenario.yaml>")
	}
	sc, err := scenario.Load(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	ad, err := adapter.New(sc.Adapter)
	if err != nil {
		log.Fatalf("adapter: %v", err)
	}

	r := &runner.Runner{Adapter: ad}

	spec := adapter.EnvSpec{
		State: adapter.StateRef{
			ID:      sc.State.ID,
			Kind:    sc.State.Kind,
			Locator: sc.State.Locator,
		},
		Config: adapter.CreateConfig{
			DBVersion: sc.Config.DBVersion,
			Resources: adapter.ResourceLimits{CPU: sc.Config.Resources.CPU, RAM: sc.Config.Resources.RAMGB},
			Isolation: sc.Config.Isolation,
			Mode:      sc.Config.Mode,
			Timeout:   time.Duration(sc.Config.TimeoutS) * time.Second,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	eRef, _, err := r.Create(ctx, spec)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Delete(ctx, eRef)

	res, err := scenario.RunS4(ctx, r, spec, eRef.ConnectionString, sc.S4.MutationSQL)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := json.MarshalIndent(res, "", "  ")
	fmt.Println(string(out))
}
