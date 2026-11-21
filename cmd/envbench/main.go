package main

import (
	"context"
	"encoding/json"
	"envbench/internal/adapter"
	_ "envbench/internal/adapter/docker" // for init()
	"envbench/internal/engine"
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

	e := &engine.Engine{Adapter: ad}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	res, err := scenario.Run(ctx, e, sc)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := json.MarshalIndent(res, "", "  ")
	fmt.Println(string(out))
}
