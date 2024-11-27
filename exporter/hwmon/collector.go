package hwmon

import (
	"fmt"
	"log/slog"
	"strconv"

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
}

func NewCollector(conf Conf) (c *collector, err error) {
	c = &collector{path: conf.Path}
	err = Init()
	if err != nil {
		return
	}
	err = lmsensors.Init()
	return
}

func (c *collector) Close() error {
	return nil
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	c.cpu_energy_desc = prometheus.NewDesc(prefix+"cpu_energy", "cpu total energy use in uj", []string{"package", "sensor"}, nil)
	c.cpu_freq_desc = prometheus.NewDesc(prefix+"cpu_frequency", "cpu frequency in KHz", []string{"package", "sensor"}, nil)
	c.fan_speed_desc = prometheus.NewDesc(prefix+"fan_speed", "fan speed from libsensors", []string{"chip", "sensor"}, nil)
	c.temp_desc = prometheus.NewDesc(prefix+"temp_celsius", "temperature from libsensors", []string{"chip", "sensor"}, nil)

	ch <- c.cpu_energy_desc
	ch <- c.cpu_freq_desc
	ch <- c.fan_speed_desc
	ch <- c.temp_desc
}

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

	sensors, err := lmsensors.Get()
	if err != nil {
		slog.Error("failed to get lmsensors data", "err", err)
		return
	}
	for _, chip := range sensors.Chips {
		for _, reading := range chip.Sensors {
			switch reading.SensorType {
			case lmsensors.Fan:
				ch <- prometheus.MustNewConstMetric(c.fan_speed_desc, prometheus.GaugeValue, reading.Value, chip.ID, reading.Name)
			case lmsensors.Temperature:
				ch <- prometheus.MustNewConstMetric(c.temp_desc, prometheus.GaugeValue, reading.Value, chip.ID, reading.Name)
			}
		}
	}
}

func (c *collector) Path() string {
	return c.path
}
