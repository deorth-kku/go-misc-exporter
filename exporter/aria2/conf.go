package aria2

type Conf struct {
	Path    string       `json:"path,omitzero"`
	Servers []ServerConf `json:"servers,omitzero"`
}

type ServerConf struct {
	Rpc     string  `json:"rpc,omitzero"`
	Secret  string  `json:"secret,omitzero"`
	Timeout float64 `json:"timeout,omitzero"`
}
