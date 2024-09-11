package hwmon

import (
	"strconv"

	"github.com/deorth-kku/go-misc-exporter/common"
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
}

func NewCollector(conf Conf) (c *collector, err error) {
	c = &collector{path: conf.Path}
	_, err = IntelRaplEnergy()
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
	c.cpu_energy_desc = prometheus.NewDesc(prefix+"cpu_energy", "cpu total energy use in uj", []string{"sensor"}, nil)
	c.fan_speed_desc = prometheus.NewDesc(prefix+"fan_speed", "fan speed from libsensors", []string{"chip", "sensor"}, nil)
	ch <- c.cpu_energy_desc
	ch <- c.fan_speed_desc
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	rapl := common.Must(IntelRaplEnergy())
	for sensor, value := range rapl {
		ch <- prometheus.MustNewConstMetric(c.cpu_energy_desc, prometheus.CounterValue, float64(value), "rapl-"+strconv.Itoa(sensor))
	}
	sensors := common.Must(lmsensors.Get())
	for _, chip := range sensors.Chips {
		for _, reading := range chip.Sensors {
			if reading.SensorType != lmsensors.Fan {
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.fan_speed_desc, prometheus.GaugeValue, reading.Value, chip.ID, reading.Name)
		}
	}
}

func (c *collector) Path() string {
	return c.path
}
