package hwmon

import (
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"sync/atomic"
	_ "unsafe"

	"github.com/mt-inside/go-lmsensors"
	"github.com/prometheus/client_golang/prometheus"
)

const prefix = "hwmon_"

type Conf struct {
	Path string `json:"path,omitempty"`
}

type collector struct {
	path            string
	cpu_energy_desc *prometheus.Desc
	fan_speed_desc  *prometheus.Desc
	temp_desc       *prometheus.Desc
	cpu_freq_desc   *prometheus.Desc
	closed          atomic.Bool
}

var (
	lmsensors_init_lock sync.Mutex
	lmsensors_init_err  error
	count               atomic.Int32
)

func init_lmsensors() error {
	if count.Load() == 0 {
		lmsensors_init_lock.Lock()
		lmsensors_init_err = lmsensors.Init()
		lmsensors_init_lock.Unlock()
	}
	count.Add(1)
	return lmsensors_init_err
}

func NewCollector(conf Conf) (c *collector, err error) {
	c = &collector{path: conf.Path}
	err = Init()
	if err != nil {
		return
	}
	err = init_lmsensors()
	return
}

func (c *collector) Close() error {
	if c.closed.Load() {
		return nil
	}
	count.Add(-1)
	if count.Load() == 0 {
		c.closed.Store(true)
		lmsensors.Cleanup()
	}
	return nil
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	c.cpu_energy_desc = prometheus.NewDesc(prefix+"cpu_energy", "cpu total energy use in uj", []string{"package", "sensor"}, nil)
	c.cpu_freq_desc = prometheus.NewDesc(prefix+"cpu_frequency", "cpu frequency in KHz", []string{"package", "sensor"}, nil)
	c.fan_speed_desc = prometheus.NewDesc(prefix+"fan_speed", "fan speed from libsensors", []string{"chip", "adapter", "sensor"}, nil)
	c.temp_desc = prometheus.NewDesc(prefix+"temp_celsius", "temperature from libsensors", []string{"chip", "adapter", "sensor", "type"}, nil)

	ch <- c.cpu_energy_desc
	ch <- c.cpu_freq_desc
	ch <- c.fan_speed_desc
	ch <- c.temp_desc
}

//go:linkname chips github.com/mt-inside/go-lmsensors.chips
func chips(func(lmsensors.ChipPtr) bool)

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	rapl, err := UseSensors()
	if err != nil {
		slog.Error("failed to get rapl energy", "err", err)
	} else {
		for pkg, value := range rapl {
			pkgstr := strconv.Itoa(pkg)
			ch <- prometheus.MustNewConstMetric(c.cpu_energy_desc, prometheus.CounterValue, float64(value.Package), pkgstr, "package")
			for core, value := range value.PerCore {
				ch <- prometheus.MustNewConstMetric(c.cpu_energy_desc, prometheus.CounterValue, float64(value), pkgstr, fmt.Sprintf("core-%d", core))
			}
		}
	}

	freqs, err := ReadCPUFreq()
	if err != nil {
		slog.Error("failed to read cpu frequency")
	} else {
		for pkg, cores := range freqs {
			pkgstr := strconv.Itoa(pkg)
			for coreid, value := range cores {
				ch <- prometheus.MustNewConstMetric(c.cpu_freq_desc, prometheus.GaugeValue, float64(value), pkgstr, fmt.Sprintf("core-%d", coreid))
			}
		}
	}

	for chip := range chips {
		chipname, adp := chip.Name(), chip.Adapter()
		for reading := range chip.Sensors {
			switch r := reading.(type) {
			case *lmsensors.FanSensor:
				ch <- prometheus.MustNewConstMetric(c.fan_speed_desc, prometheus.GaugeValue, r.Value, chipname, adp, r.Name)
			case *lmsensors.TempSensor:
				ch <- prometheus.MustNewConstMetric(c.temp_desc, prometheus.GaugeValue, r.Value, chipname, adp, r.Name, r.TempType.String())
			}
		}
	}
}

func (c *collector) Path() string {
	return c.path
}
