package hwmon

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const rapl_path = "/sys/class/powercap/intel-rapl"

func IntelRaplEnergy() (energy_uj []uint64, err error) {
	dirs, err := os.ReadDir(rapl_path)
	if err != nil {
		return
	}
	var data []byte
	var uj_int64 uint64
	for _, dir := range dirs {
		if dir.IsDir() && strings.HasPrefix(dir.Name(), "intel-rapl:") {
			uj_file := filepath.Join(rapl_path, dir.Name(), "energy_uj")
			if _, err = os.Stat(uj_file); err == nil {
				data, err = os.ReadFile(uj_file)
				if err != nil {
					continue
				}
				uj_int64, err = strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
				if err != nil {
					continue
				}
				energy_uj = append(energy_uj, uj_int64)
			} else {
				return nil, err
			}
		}
	}
	return
}
