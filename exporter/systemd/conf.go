package systemd

type Conf struct {
	Path       string   `json:"path,omitzero"`
	Properties []string `json:"properties,omitzero"`
	Units      []string `json:"units,omitzero"`
	Patterns   []string `json:"patterns,omitzero"`
	States     []string `json:"states,omitzero"`
	Timeout    float64  `json:"timeout,omitzero"`
}
