package systemd

import (
	"context"
	"iter"
	"log/slog"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/deorth-kku/go-misc-exporter/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stoewer/go-strcase"
)

var default_exported_properties = []string{
	"CPUUsageNSec",
	"MemoryCurrent",
	"TasksCurrent",
}

const (
	head      = "systemd_"
	unit_head = head + "unit_"
)

type collector struct {
	*dbus.Conn
	Conf
	descs        map[string]*prometheus.Desc
	uptime_desc  *prometheus.Desc
	loaded_count *prometheus.Desc
	failed_count *prometheus.Desc
	jobs_count   *prometheus.Desc
	version      string
	cancel       context.CancelFunc
}

func NewCollector(conf Conf) (col *collector, err error) {
	col = new(collector)
	col.Conf = conf
	if len(col.Properties) == 0 {
		col.Properties = default_exported_properties
	}
	if col.Timeout == 0 {
		col.Timeout = 10
	}
	col.descs = make(map[string]*prometheus.Desc)
	var ctx context.Context
	ctx, col.cancel = context.WithCancel(context.Background())
	timer := time.AfterFunc(common.FloatDuration(col.Timeout), col.cancel)
	defer timer.Stop()
	col.Conn, err = dbus.NewSystemdConnectionContext(ctx)
	if err != nil {
		return
	}
	col.version, err = col.GetManagerProperty("Version")
	return
}

func (c *collector) Path() string {
	return c.Conf.Path
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	c.uptime_desc = prometheus.NewDesc(head+"uptime_seconds", "systemd uptime", nil, prometheus.Labels{"version": c.version})
	c.jobs_count = prometheus.NewDesc(head+"jobs_queued_count", "systemd queued jobs count ", nil, nil)
	c.failed_count = prometheus.NewDesc(unit_head+"failed_count", "systemd failed unit count", nil, nil)
	c.loaded_count = prometheus.NewDesc(unit_head+"loaded_count", "systemd loaded unit count", nil, nil)
	ch <- c.uptime_desc
	ch <- c.jobs_count
	ch <- c.failed_count
	ch <- c.loaded_count

	for _, prop := range c.Properties {
		c.descs[prop] = prometheus.NewDesc(unit_head+strcase.SnakeCase(prop), prop, []string{"unit"}, nil)
		ch <- c.descs[prop]
	}
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := common.TimeoutContext(c.Timeout)
	defer cancel()

	uptimestr, err := c.GetManagerProperty("UserspaceTimestamp")
	if err != nil {
		slog.Error("failed to get uptime", "err", err)
	} else {
		t0, err := parseTimeStamp(uptimestr)
		if err != nil {
			slog.Error("failed to parse uptime", "string", uptimestr, "err", err)
		} else {
			ch <- prometheus.MustNewConstMetric(c.uptime_desc, prometheus.GaugeValue, time.Since(t0).Seconds())
		}
	}

	jobs, err := c.ListJobsContext(ctx)
	if err != nil {
		slog.Error("failed to list jobs", "err", err)
	} else {
		ch <- prometheus.MustNewConstMetric(c.jobs_count, prometheus.GaugeValue, float64(len(jobs)))
	}

	loaded, err := c.ListUnitsFilteredContext(ctx, []string{"loaded"})
	if err != nil {
		slog.Error("failed to list loaded units", "err", err)
	} else {
		dedup := slices.Compact(slices.Collect(unitsBaseNames(loaded)))
		ch <- prometheus.MustNewConstMetric(c.loaded_count, prometheus.GaugeValue, float64(len(dedup)))
	}

	failed, err := c.ListUnitsFilteredContext(ctx, []string{"failed"})
	if err != nil {
		slog.Error("failed to list failed units", "err", err)
	} else {
		ch <- prometheus.MustNewConstMetric(c.failed_count, prometheus.GaugeValue, float64(len(failed)))
	}

	var units iter.Seq[string]
	if len(c.Units) == 0 {
		us, err := c.ListUnitsByPatternsContext(ctx, c.States, c.Patterns)
		if err != nil {
			slog.Error("failed to list units", "err", err)
			return
		}
		units = unitsNames(us)
	} else {
		units = slices.Values(c.Units)
	}

	for unit := range units {
		props, err := c.GetAllPropertiesContext(ctx, unit)
		if err != nil {
			slog.Error("failed to get unit properties", "unit", unit, "err", err)
			continue
		}
		for prop, desc := range c.descs {
			v, ok := props[prop]
			if !ok {
				slog.Debug("unit does not have property", "unit", unit, "property", prop)
				continue
			}
			vv, ok := anyToFloat64(v)
			if !ok {
				slog.Warn("failed to convert property to float64", "property", prop)
				continue
			}
			var tt prometheus.ValueType
			if prop == "CPUUsageNSec" {
				tt = prometheus.CounterValue
			} else {
				tt = prometheus.GaugeValue
			}
			ch <- prometheus.MustNewConstMetric(desc, tt, vv, unit)
		}
	}

}

func (c *collector) Close() error {
	c.Conn.Close()
	c.cancel()
	return nil
}

func anyToFloat64(f any) (float64, bool) {
	switch v := f.(type) {
	case uint64:
		if v == math.MaxUint64 {
			return math.Inf(-1), true
		}
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	case int8:
		return float64(v), true
	case int:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}

func unitsNames(us []dbus.UnitStatus) iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, unit := range us {
			if !yield(unit.Name) {
				return
			}
		}
	}
}

func parseTimeStamp(ts string) (t time.Time, err error) {
	_, ts, _ = strings.Cut(ts, " ")
	t0, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return
	}
	return time.UnixMicro(t0), nil
}

func unitBaseName(us dbus.UnitStatus) string {
	lvs := strings.Split(us.Name, "\\")
	return lvs[len(lvs)-1]
}

func unitsBaseNames(us []dbus.UnitStatus) iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, unit := range us {
			if !yield(unitBaseName(unit)) {
				return
			}
		}
	}
}
