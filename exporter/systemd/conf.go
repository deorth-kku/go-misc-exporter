package systemd

type Conf struct {
	Path       string   `json:"path,omitempty"`
	Properties []string `json:"properties,omitempty"`
	Units      []string `json:"units,omitempty"`
	Patterns   []string `json:"patterns,omitempty"`
	States     []string `json:"states,omitempty"`
	Timeout    float64  `json:"timeout,omitempty"`
}
