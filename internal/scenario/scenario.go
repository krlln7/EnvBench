package scenario

import (
	"envbench/internal/adapter"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Scenario struct {
	ID      string       `yaml:"id"`
	Adapter string       `yaml:"adapter"`
	State   State        `yaml:"state"`
	Config  Config       `yaml:"config"`
	Run     ScenarioSpec `yaml:"scenario"`
	Repeat  int          `yaml:"repeats"`
}

type ScenarioSpec struct {
	Kind   string    `yaml:"name"` // s1, s2, ...
	Params yaml.Node `yaml:"params"`
}

type State struct {
	ID      string `yaml:"logical_id"`
	Kind    string `yaml:"kind"`
	Locator string `yaml:"locator"`
}

type Config struct {
	DBVersion string `yaml:"db_version"`
	Resources struct {
		CPU   int `yaml:"cpu"`
		RAMGB int `yaml:"ram_gb"`
	} `yaml:"resources"`
	Mode      string `yaml:"mode"`
	Isolation string `yaml:"isolation"`
	TimeoutS  int    `yaml:"timeout_s"`
}

func Load(path string) (Scenario, error) {
	var s Scenario
	b, err := os.ReadFile(path)
	if err != nil {
		return s, err
	}
	if err := yaml.Unmarshal(b, &s); err != nil {
		return s, err
	}
	if s.Repeat == 0 {
		s.Repeat = 1
	}
	return s, nil
}

func (s Scenario) ToEnvSpec() adapter.EnvSpec {
	return adapter.EnvSpec{
		State: adapter.StateRef{
			ID:      s.State.ID,
			Kind:    s.State.Kind,
			Locator: s.State.Locator,
		},
		Config: adapter.CreateConfig{
			DBVersion: s.Config.DBVersion,
			Resources: adapter.ResourceLimits{
				CPU: s.Config.Resources.CPU,
				RAM: s.Config.Resources.RAMGB,
			},
			Isolation: s.Config.Isolation,
			Mode:      s.Config.Mode,
			Timeout:   time.Duration(s.Config.TimeoutS) * time.Second,
		},
	}
}
