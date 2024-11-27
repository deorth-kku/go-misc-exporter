package hwmon

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/deorth-kku/go-common"
)

const (
	rapl_path   = "/sys/class/powercap/intel-rapl"
	rapl_head   = "intel-rapl:"
	uj_basefile = "energy_uj"

	UNIT_MSR             = 0xC0010299
	CORE_MSR             = 0xC001029A
	PACKAGE_MSR          = 0xC001029B
	ENERGY_UNIT_MASK     = 0x1F00
	MICRO_JOULE_IN_JOULE = 1e6
)

type packagedata[T any] struct {
	Package T
	PerCore []T
}

type packageFiles packagedata[string]
type PackageEnergy packagedata[uint64]

var (
	intel_rapl_files []packageFiles
	amd_msr_files    []packageFiles
	UseSensors       func() ([]PackageEnergy, error)
)

func Init() error {
	amd, err := detect_amd_msr()
	if err == nil {
		amd_msr_files = amd
		UseSensors = AMDMSREnergy
		return nil
	}
	intel, err := detect_intel_rapl()
	if err == nil {
		intel_rapl_files = intel
		UseSensors = IntelRaplEnergy
	}
	return err
}

func is_file(file string) error {
	stat, err := os.Stat(file)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return fmt.Errorf("%s is a dir", file)
	}
	return nil
}

func detect_intel_rapl() (files []packageFiles, err error) {
	dirs, err := os.ReadDir(rapl_path)
	if err != nil {
		return
	}
	for _, dir := range dirs {
		if dir.IsDir() && strings.HasPrefix(dir.Name(), rapl_head) {
			var pkg packageFiles
			pkg_dir := filepath.Join(rapl_path, dir.Name())
			pkg.Package = filepath.Join(pkg_dir, uj_basefile)
			err = is_file(pkg.Package)
			if err != nil {
				return
			}
			dirs, err = os.ReadDir(pkg_dir)
			if err != nil {
				return
			}
			for _, dir := range dirs {
				if dir.IsDir() && strings.HasPrefix(dir.Name(), rapl_head) {
					file := filepath.Join(pkg_dir, dir.Name(), uj_basefile)
					err = is_file(file)
					if err != nil {
						return
					}
					pkg.PerCore = append(pkg.PerCore, file)
				}
			}
			files = append(files, pkg)
		}
	}
	return
}

func IntelRaplEnergy() (energy_uj []PackageEnergy, err error) {
	if len(intel_rapl_files) == 0 {
		return nil, errors.New("intel rapl not initialize")
	}
	energy_uj = make([]PackageEnergy, len(intel_rapl_files))
	for i, pkg := range intel_rapl_files {
		energy_uj[i].Package, err = readFileAsUint(pkg.Package)
		if err != nil {
			return
		}
		energy_uj[i].PerCore = make([]uint64, len(pkg.PerCore))
		for j, core := range pkg.PerCore {
			energy_uj[i].PerCore[j], err = readFileAsUint(core)
			if err != nil {
				return
			}
		}
	}
	return
}

func readFileAsUint(file string) (uint64, error) {
	if _, err := os.Stat(file); err == nil {
		data, err := os.ReadFile(file)
		if err != nil {
			return 0, err
		}
		return strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
	} else {
		return 0, err
	}
}

func readMSR(file string, offsets ...int64) (data []uint64, err error) {
	f, err := os.Open(file)
	if err != nil {
		return
	}
	defer f.Close()
	data = make([]uint64, len(offsets))
	for i, offset := range offsets {
		_, err = f.Seek(offset, 0)
		if err != nil {
			return
		}
		raw := make([]byte, 8)
		_, err = f.Read(raw)
		if err != nil {
			return
		}
		data[i] = binary.LittleEndian.Uint64(raw)
	}
	return
}

func energyFactor(unit_msr uint64) uint64 {
	return (unit_msr & ENERGY_UNIT_MASK) >> 8
}

func readCoreEnergyUj(file string) (uint64, error) {
	data, err := readMSR(file, UNIT_MSR, CORE_MSR)
	if err != nil {
		return 0, err
	}
	return energyFactor(data[0]) * data[1], nil
}

func readPackageEnergyUj(file string) (uint64, error) {
	data, err := readMSR(file, UNIT_MSR, PACKAGE_MSR)
	if err != nil {
		return 0, err
	}
	return energyFactor(data[0]) * data[1], nil
}

var (
	cpu_struct      map[uint64]common.Set[uint64]
	cpu_struct_err  error
	cpu_struct_once sync.Once
)

func detect_cpu_struct() (map[uint64]common.Set[uint64], error) {
	cpu_struct_once.Do(func() {
		cpu_struct = make(map[uint64]common.Set[uint64])
		var pkgid, cid uint64
		for i := range runtime.NumCPU() {
			pkgid, cpu_struct_err = readFileAsUint(fmt.Sprintf("/sys/devices/system/cpu/cpu%d/topology/physical_package_id", i))
			if cpu_struct_err != nil {
				return
			}
			set, ok := cpu_struct[pkgid]
			if !ok {
				set = common.NewSet[uint64]()
				cpu_struct[pkgid] = set
			}
			cid, cpu_struct_err = readFileAsUint(fmt.Sprintf("/sys/devices/system/cpu/cpu%d/topology/core_id", i))
			if !set.Has(cid) {
				set.Add(cid)
			}
		}
	})
	return cpu_struct, cpu_struct_err
}

func detect_amd_msr() (files []packageFiles, err error) {
	detect_cpu_struct()
	if cpu_struct_err != nil {
		return nil, cpu_struct_err
	}
	files = make([]packageFiles, len(cpu_struct))
	for i := range files {
		pkg, ok := cpu_struct[uint64(i)]
		if !ok {
			err = fmt.Errorf("missing package %d", i)
			return
		}
		if pkg.Len() == 0 {
			err = fmt.Errorf("no core for package %d", i)
			return
		}
		pkg_slice := pkg.Slice()
		slices.Sort(pkg_slice)
		files[i].Package = fmt.Sprintf("/dev/cpu/%d/msr", pkg_slice[0])
		err = is_file(files[i].Package)
		if err != nil {
			return
		}
		files[i].PerCore = make([]string, len(pkg_slice))
		for ci, c := range pkg_slice {
			files[i].PerCore[ci] = fmt.Sprintf("/dev/cpu/%d/msr", c)
			err = is_file(files[i].PerCore[ci])
			if err != nil {
				return
			}
		}
	}
	return
}

func AMDMSREnergy() (e []PackageEnergy, err error) {
	if len(amd_msr_files) == 0 {
		err = errors.New("amd msr not initialize")
		return
	}
	e = make([]PackageEnergy, len(amd_msr_files))
	for i, pkg := range amd_msr_files {
		e[i].Package, err = readPackageEnergyUj(pkg.Package)
		if err != nil {
			return
		}
		e[i].PerCore = make([]uint64, len(pkg.PerCore))
		for j, core := range pkg.PerCore {
			e[i].PerCore[j], err = readCoreEnergyUj(core)
			if err != nil {
				return
			}
		}
	}
	return
}
