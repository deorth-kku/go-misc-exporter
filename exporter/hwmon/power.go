package hwmon

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	rapl_path   = "/sys/class/powercap/intel-rapl"
	rapl_head   = "intel-rapl:"
	uj_basefile = "energy_uj"
)

type RaplPackageEnergy struct {
	Package uint64
	PerCore []uint64
}

func IntelRaplEnergy() (energy_uj []RaplPackageEnergy, err error) {
	dirs, err := os.ReadDir(rapl_path)
	if err != nil {
		return
	}
	for _, dir := range dirs {
		if dir.IsDir() && strings.HasPrefix(dir.Name(), rapl_head) {
			var pkg RaplPackageEnergy
			pkg_dir := filepath.Join(rapl_path, dir.Name())
			uj_file := filepath.Join(pkg_dir, uj_basefile)
			pkg.Package, err = readRaplFile(uj_file)
			if err != nil {
				return
			}

			dirs, err = os.ReadDir(pkg_dir)
			if err != nil {
				return
			}
			for _, dir := range dirs {
				var core uint64
				if dir.IsDir() && strings.HasPrefix(dir.Name(), rapl_head) {
					uj_file := filepath.Join(pkg_dir, dir.Name(), uj_basefile)
					core, err = readRaplFile(uj_file)
					if err != nil {
						return
					}
					pkg.PerCore = append(pkg.PerCore, core)
				}
			}
			energy_uj = append(energy_uj, pkg)
		}
	}
	return
}

func readRaplFile(uj_file string) (uint64, error) {
	if _, err := os.Stat(uj_file); err == nil {
		data, err := os.ReadFile(uj_file)
		if err != nil {
			return 0, err
		}
		return strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
	} else {
		return 0, err
	}
}
