package zfs

import (
	"slices"

	"git.dolansoft.org/lorenz/go-zfs/ioctl"
)

type Conf struct {
	Skip   []string `json:"skip,omitzero"`
	Path   string   `json:"path,omitzero"`
	Device string   `json:"device,omitzero"`
}

func (c Conf) ListPools() ([]string, error) {
	configs, err := ioctl.PoolConfigs()
	if err != nil {
		return nil, err
	}
	pools := make([]string, 0, len(configs))
	for pool := range configs {
		if c.MatchSkip(pool) {
			continue
		}
		pools = append(pools, pool)
	}
	return pools, nil
}

func (c Conf) MatchSkip(pool string) bool {
	return slices.Contains(c.Skip, pool)
}
