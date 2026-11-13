package scenario

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Scenario struct {
	ID      string `yaml:"id"`
	Adapter string `yaml:"adapter"`
	State   State  `yaml:"state"`
	Config  Config `yaml:"config"`
	S4      *S4    `yaml:"s4,omitempty"`
	Repeat  int    `yaml:"repeats"`
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
