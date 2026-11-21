package result

type Result struct {
	ScenarioID string            `json:"scenario_id"`
	Adapter    string            `json:"adapter"`
	Verdict    string            `json:"verdict"` // correctness_failure, creation_failure
	Checks     map[string]string `json:"checks"`
	Metrics    map[string]any    `json:"metrics,omitempty"`
}
