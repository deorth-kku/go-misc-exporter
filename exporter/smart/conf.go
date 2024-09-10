package smart

type Conf struct {
	Skip []string `json:"skip,omitempty"`
	Path string   `json:"path,omitempty"`
}
