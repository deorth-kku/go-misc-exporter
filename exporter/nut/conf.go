package nut

type Server struct {
	Host     string  `json:"host,omitzero"`
	Port     uint16  `json:"port,omitzero"`
	Username string  `json:"username,omitzero"`
	Password string  `json:"password,omitzero"`
	Timeout  float64 `json:"timeout,omitzero"`
}

type Conf struct {
	Servers []Server `json:"servers,omitzero"`
	Path    string   `json:"path,omitzero"`
}
