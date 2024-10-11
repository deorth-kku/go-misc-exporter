package ryzenadj

import (
	"fmt"
	"log/slog"
	"runtime"

	"github.com/deorth-kku/go-common"
	"github.com/deorth-kku/ryzenadj-go/lib"
	"github.com/prometheus/client_golang/prometheus"
)

type Conf struct {
	Path string `json:"path,omitempty"`
}

type getFunc func() float32
type getCoreFunc func(uint32) float32

type collector struct {
	lib.RyzenAccess
	Conf
	infoDesc     *prometheus.Desc
	getDescs     common.PairSlice[*prometheus.Desc, getFunc]
	getCoreDescs common.PairSlice[*prometheus.Desc, getCoreFunc]
}

var corecount = uint32(runtime.NumCPU())

func NewCollector(conf Conf) (c *collector, err error) {
	c = &collector{Conf: conf}
	c.RyzenAccess, err = lib.NewRyzenAccess()
	return
}

func (c *collector) Path() string {
	return c.Conf.Path
}

func (c *collector) Close() error {
	c.Cleanup()
	return nil
}

const head = "ryzenadj_"

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	c.infoDesc = prometheus.NewDesc(head+"platform_info", "", nil, prometheus.Labels{
		"family":             c.GetCpuFamily().String(),
		"bios_interface_ver": fmt.Sprintf("0x%04X", c.GetBiosIfVer()),
	})
	getMap := common.PairSlice[string, getFunc]{
		{Key: "stapm_limit", Value: c.GetStapmLimit},
		{Key: "stapm_value", Value: c.GetStapmValue},
		{Key: "fast_limit", Value: c.GetFastLimit},
		{Key: "fast_value", Value: c.GetFastValue},
		{Key: "slow_limit", Value: c.GetSlowLimit},
		{Key: "slow_value", Value: c.GetSlowValue},
		{Key: "apu_slow_limit", Value: c.GetApuSlowLimit},
		{Key: "apu_slow_value", Value: c.GetApuSlowValue},
		{Key: "vrm_current", Value: c.GetVrmCurrent},
		{Key: "vrm_current_value", Value: c.GetVrmCurrentValue},
		{Key: "vrmsoc_current", Value: c.GetVrmsocCurrent},
		{Key: "vrmsoc_current_value", Value: c.GetVrmsocCurrentValue},
		{Key: "vrmmax_current", Value: c.GetVrmmaxCurrent},
		{Key: "vrmmax_current_value", Value: c.GetVrmmaxCurrentValue},
		{Key: "vrmsocmax_current", Value: c.GetVrmsocmaxCurrent},
		{Key: "vrmsocmax_current_value", Value: c.GetVrmsocmaxCurrentValue},
		{Key: "tctl_temp", Value: c.GetTctlTemp},
		{Key: "tctl_temp_value", Value: c.GetTctlTempValue},
		{Key: "apu_skin_temp_limit", Value: c.GetApuSkinTempLimit},
		{Key: "apu_skin_temp_value", Value: c.GetApuSkinTempValue},
		{Key: "dgpu_skin_temp_limit", Value: c.GetDgpuSkinTempLimit},
		{Key: "dgpu_skin_temp_value", Value: c.GetDgpuSkinTempValue},
		{Key: "psi0_current", Value: c.GetPsi0Current},
		{Key: "psi0soc_current", Value: c.GetPsi0socCurrent},
		{Key: "stapm_time", Value: c.GetStapmTime},
		{Key: "slow_time", Value: c.GetSlowTime},
		{Key: "cclk_setpoint", Value: c.GetCclkSetpoint},
		{Key: "cclk_busy_value", Value: c.GetCclkBusyValue},
		{Key: "l3_clk", Value: c.GetL3Clk},
		{Key: "l3_logic", Value: c.GetL3Logic},
		{Key: "l3_vddm", Value: c.GetL3Vddm},
		{Key: "l3_temp", Value: c.GetL3Temp},
		{Key: "gfx_clk", Value: c.GetGfxClk},
		{Key: "gfx_temp", Value: c.GetGfxTemp},
		{Key: "gfx_volt", Value: c.GetGfxVolt},
		{Key: "mem_clk", Value: c.GetMemClk},
		{Key: "fclk", Value: c.GetFclk},
		{Key: "soc_power", Value: c.GetSocPower},
		{Key: "soc_volt", Value: c.GetSocVolt},
		{Key: "socket_power", Value: c.GetSocketPower},
	}
	for name, get := range getMap.Range {
		if common.IsNaN(get()) {
			continue
		}
		c.getDescs = append(c.getDescs, common.Pair[*prometheus.Desc, getFunc]{
			Key:   prometheus.NewDesc(head+name, "", nil, nil),
			Value: get,
		})
	}

	core := []string{"core"}
	getCoreMap := common.PairSlice[string, getCoreFunc]{
		{Key: "core_clk", Value: c.GetCoreClk},
		{Key: "core_volt", Value: c.GetCoreVolt},
		{Key: "core_power", Value: c.GetCorePower},
		{Key: "core_temp", Value: c.GetCoreTemp},
	}
	for name, get := range getCoreMap.Range {
		if common.IsNaN(get(0)) {
			continue
		}
		c.getCoreDescs = append(c.getCoreDescs, common.Pair[*prometheus.Desc, getCoreFunc]{
			Key:   prometheus.NewDesc(head+name, "", core, nil),
			Value: get,
		})
	}
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	err := c.Refresh()
	if err != nil {
		slog.Error("failed to refresh ryzenadj", "err", err)
		return
	}
	ch <- prometheus.MustNewConstMetric(c.infoDesc, prometheus.GaugeValue, 0)

	for desc, get := range c.getDescs.Range {
		ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, float64(get()))
	}
	for desc, get := range c.getCoreDescs.Range {
		for c := range corecount {
			ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, float64(get(c)), fmt.Sprint(c))
		}
	}
}
