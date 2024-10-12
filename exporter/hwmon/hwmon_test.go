package hwmon

import (
	"fmt"
	"testing"

	"github.com/deorth-kku/go-common"
	"github.com/deorth-kku/go-misc-exporter/cmd"
)

var _ cmd.Collector = new(collector)

func TestRapl(t *testing.T) {
	var err error
	intel_rapl_files, err = detect_intel_rapl()
	if err != nil {
		t.Error(err)
		return
	}
	rapl, err := IntelRaplEnergy()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(rapl)
}

func TestMSR(t *testing.T) {
	var err error
	amd_msr_files, err = detect_amd_msr()
	if err != nil {
		t.Error(err)
		return
	}
	data, err := AMDMSREnergy()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(data)
}

func TestCollector(t *testing.T) {
	err := cmd.TestCollectorThenClose(common.Must(NewCollector(Conf{})))
	if err != nil {
		t.Error(err)
		return
	}
}
