package smart

import (
	"path/filepath"
)

type Conf struct {
	Skip []string `json:"skip,omitzero"`
	Path string   `json:"path,omitzero"`
}

func (c Conf) MatchSkip(name string) bool {
	for i, skip := range c.Skip {
		if filepath.IsAbs(skip) {
			if real, err := filepath.EvalSymlinks(skip); err == nil {
				skip = real
			}
			c.Skip[i] = filepath.Base(skip)
		}
		if name == c.Skip[i] {
			return true
		}
	}
	return false
}
