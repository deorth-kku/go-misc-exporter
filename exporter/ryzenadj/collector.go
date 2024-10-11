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
	infoDesc   *prometheus.Desc
	getMap     common.PairSlice[*prometheus.Desc, getFunc]
	getCoreMap common.PairSlice[*prometheus.Desc, getCoreFunc]
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
	c.getMap = common.PairSlice[*prometheus.Desc, getFunc]{
		{Key: prometheus.NewDesc(head+"stapm_limit", "", nil, nil), Value: c.GetStapmLimit},
		{Key: prometheus.NewDesc(head+"stapm_value", "", nil, nil), Value: c.GetStapmValue},
		{Key: prometheus.NewDesc(head+"fast_limit", "", nil, nil), Value: c.GetFastLimit},
		{Key: prometheus.NewDesc(head+"fast_value", "", nil, nil), Value: c.GetFastValue},
		{Key: prometheus.NewDesc(head+"slow_limit", "", nil, nil), Value: c.GetSlowLimit},
		{Key: prometheus.NewDesc(head+"slow_value", "", nil, nil), Value: c.GetSlowValue},
		{Key: prometheus.NewDesc(head+"apu_slow_limit", "", nil, nil), Value: c.GetApuSlowLimit},
		{Key: prometheus.NewDesc(head+"apu_slow_value", "", nil, nil), Value: c.GetApuSlowValue},
		{Key: prometheus.NewDesc(head+"vrm_current", "", nil, nil), Value: c.GetVrmCurrent},
		{Key: prometheus.NewDesc(head+"vrm_current_value", "", nil, nil), Value: c.GetVrmCurrentValue},
		{Key: prometheus.NewDesc(head+"vrmsoc_current", "", nil, nil), Value: c.GetVrmsocCurrent},
		{Key: prometheus.NewDesc(head+"vrmsoc_current_value", "", nil, nil), Value: c.GetVrmsocCurrentValue},
		{Key: prometheus.NewDesc(head+"vrmmax_current", "", nil, nil), Value: c.GetVrmmaxCurrent},
		{Key: prometheus.NewDesc(head+"vrmmax_current_value", "", nil, nil), Value: c.GetVrmmaxCurrentValue},
		{Key: prometheus.NewDesc(head+"vrmsocmax_current", "", nil, nil), Value: c.GetVrmsocmaxCurrent},
		{Key: prometheus.NewDesc(head+"vrmsocmax_current_value", "", nil, nil), Value: c.GetVrmsocmaxCurrentValue},
		{Key: prometheus.NewDesc(head+"tctl_temp", "", nil, nil), Value: c.GetTctlTemp},
		{Key: prometheus.NewDesc(head+"tctl_temp_value", "", nil, nil), Value: c.GetTctlTempValue},
		{Key: prometheus.NewDesc(head+"apu_skin_temp_limit", "", nil, nil), Value: c.GetApuSkinTempLimit},
		{Key: prometheus.NewDesc(head+"apu_skin_temp_value", "", nil, nil), Value: c.GetApuSkinTempValue},
		{Key: prometheus.NewDesc(head+"dgpu_skin_temp_limit", "", nil, nil), Value: c.GetDgpuSkinTempLimit},
		{Key: prometheus.NewDesc(head+"dgpu_skin_temp_value", "", nil, nil), Value: c.GetDgpuSkinTempValue},
		{Key: prometheus.NewDesc(head+"psi0_current", "", nil, nil), Value: c.GetPsi0Current},
		{Key: prometheus.NewDesc(head+"psi0soc_current", "", nil, nil), Value: c.GetPsi0socCurrent},
		{Key: prometheus.NewDesc(head+"stapm_time", "", nil, nil), Value: c.GetStapmTime},
		{Key: prometheus.NewDesc(head+"slow_time", "", nil, nil), Value: c.GetSlowTime},
		{Key: prometheus.NewDesc(head+"cclk_setpoint", "", nil, nil), Value: c.GetCclkSetpoint},
		{Key: prometheus.NewDesc(head+"cclk_busy_value", "", nil, nil), Value: c.GetCclkBusyValue},
		{Key: prometheus.NewDesc(head+"l3_clk", "", nil, nil), Value: c.GetL3Clk},
		{Key: prometheus.NewDesc(head+"l3_logic", "", nil, nil), Value: c.GetL3Logic},
		{Key: prometheus.NewDesc(head+"l3_vddm", "", nil, nil), Value: c.GetL3Vddm},
		{Key: prometheus.NewDesc(head+"l3_temp", "", nil, nil), Value: c.GetL3Temp},
		{Key: prometheus.NewDesc(head+"gfx_clk", "", nil, nil), Value: c.GetGfxClk},
		{Key: prometheus.NewDesc(head+"gfx_temp", "", nil, nil), Value: c.GetGfxTemp},
		{Key: prometheus.NewDesc(head+"gfx_volt", "", nil, nil), Value: c.GetGfxVolt},
		{Key: prometheus.NewDesc(head+"mem_clk", "", nil, nil), Value: c.GetMemClk},
		{Key: prometheus.NewDesc(head+"fclk", "", nil, nil), Value: c.GetFclk},
		{Key: prometheus.NewDesc(head+"soc_power", "", nil, nil), Value: c.GetSocPower},
		{Key: prometheus.NewDesc(head+"soc_volt", "", nil, nil), Value: c.GetSocVolt},
		{Key: prometheus.NewDesc(head+"socket_power", "", nil, nil), Value: c.GetSocketPower},
	}

	core := []string{"core"}
	c.getCoreMap = common.PairSlice[*prometheus.Desc, getCoreFunc]{
		{Key: prometheus.NewDesc(head+"core_clk", "", core, nil), Value: c.GetCoreClk},
		{Key: prometheus.NewDesc(head+"core_volt", "", core, nil), Value: c.GetCoreVolt},
		{Key: prometheus.NewDesc(head+"core_power", "", core, nil), Value: c.GetCorePower},
		{Key: prometheus.NewDesc(head+"core_temp", "", core, nil), Value: c.GetCoreTemp},
	}
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	err := c.Refresh()
	if err != nil {
		slog.Error("failed to refresh ryzenadj", "err", err)
		return
	}
	ch <- prometheus.MustNewConstMetric(c.infoDesc, prometheus.GaugeValue, 0)

	for desc, get := range c.getMap.Range {
		ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, float64(get()))
	}
	for desc, get := range c.getCoreMap.Range {
		for c := range corecount {
			ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, float64(get(c)), fmt.Sprint(c))
		}
	}
}
