package nut

type Server struct {
	Host     string  `json:"host,omitempty"`
	Port     uint16  `json:"port,omitempty"`
	Username string  `json:"username,omitempty"`
	Password string  `json:"password,omitempty"`
	Timeout  float64 `json:"timeout,omitempty"`
}

type Conf struct {
	Servers []Server `json:"servers,omitempty"`
	Path    string   `json:"path,omitempty"`
}
