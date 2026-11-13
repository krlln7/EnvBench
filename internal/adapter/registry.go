// internal/adapter/registry.go
package adapter

import (
	"fmt"
	"sort"
)

type Constructor func() (DatabaseAdapter, error)

var registry = map[string]Constructor{}

func Register(name string, c Constructor) {
	if _, ok := registry[name]; ok {
		panic("adapter already registered: " + name)
	}
	registry[name] = c
}

func New(name string) (DatabaseAdapter, error) {
	c, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown adapter %q (registered: %v)", name, Names())
	}
	return c()
}

func Names() []string {
	names := make([]string, 0, len(registry))
	for n := range registry {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
