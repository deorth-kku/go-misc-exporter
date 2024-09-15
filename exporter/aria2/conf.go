package aria2

type Conf struct {
	Path    string       `json:"path,omitempty"`
	Servers []ServerConf `json:"servers,omitempty"`
}

type ServerConf struct {
	Rpc     string  `json:"rpc,omitempty"`
	Secret  string  `json:"secret,omitempty"`
	Timeout float64 `json:"timeout,omitempty"`
}
