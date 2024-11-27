package hwmon

import (
	"fmt"
	"slices"
)

func ReadCPUFreq() ([][]uint64, error) {
	detect_cpu_struct()
	if cpu_struct_err != nil {
		return nil, cpu_struct_err
	}
	freqs := make([][]uint64, len(cpu_struct))
	for i := range freqs {
		coreids, ok := cpu_struct[uint64(i)]
		if !ok {
			return nil, fmt.Errorf("missing package %d", i)
		}
		coreids_slice := coreids.Slice()
		slices.Sort(coreids_slice)
		var corefreq uint64
		var err error
		freqs[i] = make([]uint64, coreids.Len())
		for j, coreid := range coreids_slice {
			corefreq, err = readFileAsUint(fmt.Sprintf("/sys/devices/system/cpu/cpu%d/cpufreq/scaling_cur_freq", coreid))
			if err != nil {
				return nil, err
			}
			freqs[i][j] = corefreq
		}
	}
	return freqs, nil
}
