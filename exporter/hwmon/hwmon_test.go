package hwmon

import (
	"fmt"
	"testing"

	"github.com/deorth-kku/go-misc-exporter/cmd"
	"github.com/deorth-kku/go-misc-exporter/common"
)

var _ cmd.Collector = new(collector)

func TestRapl(t *testing.T) {
	rapl, err := IntelRaplEnergy()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(rapl)
}

func TestCollector(t *testing.T) {
	err := cmd.TestCollectorThenClose(common.Must(NewCollector(Conf{})))
	if err != nil {
		t.Error(err)
		return
	}
}
