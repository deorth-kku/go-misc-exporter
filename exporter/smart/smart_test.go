package smart

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/anatol/smart.go"
	"github.com/deorth-kku/go-common"
	"github.com/deorth-kku/go-misc-exporter/cmd"
)

var _ cmd.Collector = new(collector)

func TestCollector(t *testing.T) {
	col, err := NewCollector(Conf{})
	if err != nil {
		t.Error(err)
		return
	}
	err = cmd.TestCollectorThenClose(col)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestNvme(t *testing.T) {
	path := "/dev/nvme0n1"
	dev, err := smart.OpenNVMe(path)
	if err != nil {
		t.Error(err)
		return
	}
	defer dev.Close()
	d := NewNvmeDev(path, dev)
	for _, a := range d.GetMetrics() {
		fmt.Println(a)
	}
}

func TestSata(t *testing.T) {
	path := "/dev/sda"
	dev, err := smart.OpenSata(path)
	if err != nil {
		t.Error(err)
		return
	}
	defer dev.Close()
	d := NewSataDev(path, dev)
	d.ListMetrics()
	for _, a := range d.GetMetrics() {
		fmt.Println(a)
	}
}

func TestScsi(t *testing.T) {
	path := "/dev/sde"
	dev, err := smart.OpenScsi(path)
	if err != nil {
		t.Error(err)
	}
	defer dev.Close()
	ij, _ := dev.Inquiry()
	j, _ := json.Marshal(ij)
	fmt.Println(string(j))
}

func TestParseSpinUpTime(t *testing.T) {
	// See https://github.com/netdata/netdata/issues/5919#issuecomment-487087591
	cur, avg := ParseSpinUpTime(38684000679)
	if cur == 423 && avg == 447 {

	} else {
		t.Errorf("incorrect value %d %d", cur, avg)
	}
}

func TestMatchSkip(t *testing.T) {
	const link = "/tmp/testlink"
	common.Must0(os.Symlink("/dev/null", link))
	cfg := Conf{
		Skip: []string{
			link,
		},
	}
	if !cfg.MatchSkip("null") {
		t.Error("failed to match link")
	}
	os.Remove(link)
}
