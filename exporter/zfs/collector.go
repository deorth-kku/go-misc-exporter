package zfs

import (
	"fmt"
	"log/slog"

	"git.dolansoft.org/lorenz/go-zfs/ioctl"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	head = "zfs_"

	idxVDevState          = 1
	idxVDevAlloc          = 3
	idxVDevSpace          = 4
	idxVDevOpsRead        = 9
	idxVDevOpsWrite       = 10
	idxVDevBytesRead      = 15
	idxVDevBytesWrite     = 16
	idxVDevReadErrors     = 20
	idxVDevWriteErrors    = 21
	idxVDevChecksumErrors = 22
)

type collector struct {
	Conf
	poolInfoDesc     *prometheus.Desc
	poolStateDesc    *prometheus.Desc
	poolTXGDesc      *prometheus.Desc
	vdevInfoDesc     *prometheus.Desc
	vdevStateDesc    *prometheus.Desc
	vdevAllocDesc    *prometheus.Desc
	vdevSizeDesc     *prometheus.Desc
	vdevOpsDesc      *prometheus.Desc
	vdevBytesDesc    *prometheus.Desc
	vdevReadErrDesc  *prometheus.Desc
	vdevWriteErrDesc *prometheus.Desc
	vdevCksumErrDesc *prometheus.Desc
}

func NewCollector(conf Conf) (*collector, error) {
	if err := ioctl.Init(conf.Device); err != nil {
		return nil, err
	}
	_, err := conf.ListPools()
	if err != nil {
		return nil, err
	}
	return &collector{Conf: conf}, nil
}

func (c *collector) Path() string {
	return c.Conf.Path
}

func (c *collector) Close() error {
	return nil
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	c.poolInfoDesc = prometheus.NewDesc(head+"pool_info", "zpool info", []string{"pool", "hostname"}, nil)
	c.poolStateDesc = prometheus.NewDesc(head+"pool_state", "zpool state (one metric per state label)", []string{"pool", "state"}, nil)
	c.poolTXGDesc = prometheus.NewDesc(head+"pool_txg", "latest zpool txg", []string{"pool"}, nil)
	c.vdevInfoDesc = prometheus.NewDesc(head+"vdev_info", "vdev info", []string{"pool", "vdev", "vdev_type", "vdev_class", "parent"}, nil)
	c.vdevStateDesc = prometheus.NewDesc(head+"vdev_state", "vdev state (one metric per state label)", []string{"pool", "vdev", "state"}, nil)
	c.vdevAllocDesc = prometheus.NewDesc(head+"vdev_allocated_bytes", "allocated bytes on vdev", []string{"pool", "vdev"}, nil)
	c.vdevSizeDesc = prometheus.NewDesc(head+"vdev_size_bytes", "total bytes on vdev", []string{"pool", "vdev"}, nil)
	c.vdevOpsDesc = prometheus.NewDesc(head+"vdev_ops_total", "vdev io ops", []string{"pool", "vdev", "direction"}, nil)
	c.vdevBytesDesc = prometheus.NewDesc(head+"vdev_bytes_total", "vdev io bytes", []string{"pool", "vdev", "direction"}, nil)
	c.vdevReadErrDesc = prometheus.NewDesc(head+"vdev_read_errors_total", "vdev read errors", []string{"pool", "vdev"}, nil)
	c.vdevWriteErrDesc = prometheus.NewDesc(head+"vdev_write_errors_total", "vdev write errors", []string{"pool", "vdev"}, nil)
	c.vdevCksumErrDesc = prometheus.NewDesc(head+"vdev_checksum_errors_total", "vdev checksum errors", []string{"pool", "vdev"}, nil)

	ch <- c.poolInfoDesc
	ch <- c.poolStateDesc
	ch <- c.poolTXGDesc
	ch <- c.vdevInfoDesc
	ch <- c.vdevStateDesc
	ch <- c.vdevAllocDesc
	ch <- c.vdevSizeDesc
	ch <- c.vdevOpsDesc
	ch <- c.vdevBytesDesc
	ch <- c.vdevReadErrDesc
	ch <- c.vdevWriteErrDesc
	ch <- c.vdevCksumErrDesc
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	pools, err := c.ListPools()
	if err != nil {
		slog.Error("failed to list zpools", "err", err)
		return
	}
	seen := make(map[string]struct{})
	for _, pool := range pools {
		stats, err := ioctl.PoolStats(pool)
		if err != nil {
			slog.Error("failed to collect zpool stats", "pool", pool, "err", err)
			continue
		}
		c.collectPool(ch, pool, stats, seen)
	}
}

func (c *collector) collectPool(ch chan<- prometheus.Metric, pool string, stats map[string]any, seen map[string]struct{}) {
	hostname, _ := stats["hostname"].(string)
	ch <- prometheus.MustNewConstMetric(c.poolInfoDesc, prometheus.GaugeValue, 1, pool, hostname)

	if txg, ok := uintFromAny(stats["txg"]); ok {
		ch <- prometheus.MustNewConstMetric(c.poolTXGDesc, prometheus.GaugeValue, float64(txg), pool)
	}

	var state uint64
	if v, ok := uintFromAny(stats["state"]); ok {
		state = v
	}
	ch <- prometheus.MustNewConstMetric(c.poolStateDesc, prometheus.GaugeValue, 1, pool, stateToString(state))

	vdev, ok := mapFromAny(stats["vdev_tree"])
	if !ok {
		return
	}
	c.collectVDev(ch, pool, "", "root", vdev, seen)
}

func (c *collector) collectVDev(ch chan<- prometheus.Metric, pool, parent, vdevClass string, vdev map[string]any, seen map[string]struct{}) {
	vdevType, _ := vdev["type"].(string)
	vdevName := vdevLabel(vdev, vdevType)
	vdevKey := pool + "\x00" + parent + "\x00" + vdevType + "\x00" + vdevName
	if _, ok := seen[vdevKey]; ok {
		return
	}
	seen[vdevKey] = struct{}{}

	if v, ok := uintFromAny(vdev["is_log"]); ok && v != 0 {
		if vdevClass == "data" {
			vdevClass = "log"
		}
	}
	ch <- prometheus.MustNewConstMetric(c.vdevInfoDesc, prometheus.GaugeValue, 1, pool, vdevName, vdevType, vdevClass, parent)

	stats := uintSliceFromAny(vdev["vdev_stats"])

	if state, ok := uintFromAny(vdev["state"]); ok {
		ch <- prometheus.MustNewConstMetric(c.vdevStateDesc, prometheus.GaugeValue, 1, pool, vdevName, stateToString(state))
	} else if state, ok := uintAt(stats, idxVDevState); ok {
		ch <- prometheus.MustNewConstMetric(c.vdevStateDesc, prometheus.GaugeValue, 1, pool, vdevName, stateToString(state))
	}

	if alloc, ok := uintAt(stats, idxVDevAlloc); ok {
		ch <- prometheus.MustNewConstMetric(c.vdevAllocDesc, prometheus.GaugeValue, float64(alloc), pool, vdevName)
	}
	if size, ok := uintAt(stats, idxVDevSpace); ok {
		ch <- prometheus.MustNewConstMetric(c.vdevSizeDesc, prometheus.GaugeValue, float64(size), pool, vdevName)
	}

	if opsRead, ok := uintAt(stats, idxVDevOpsRead); ok {
		ch <- prometheus.MustNewConstMetric(c.vdevOpsDesc, prometheus.CounterValue, float64(opsRead), pool, vdevName, "read")
	}
	if opsWrite, ok := uintAt(stats, idxVDevOpsWrite); ok {
		ch <- prometheus.MustNewConstMetric(c.vdevOpsDesc, prometheus.CounterValue, float64(opsWrite), pool, vdevName, "write")
	}
	if bytesRead, ok := uintAt(stats, idxVDevBytesRead); ok {
		ch <- prometheus.MustNewConstMetric(c.vdevBytesDesc, prometheus.CounterValue, float64(bytesRead), pool, vdevName, "read")
	}
	if bytesWrite, ok := uintAt(stats, idxVDevBytesWrite); ok {
		ch <- prometheus.MustNewConstMetric(c.vdevBytesDesc, prometheus.CounterValue, float64(bytesWrite), pool, vdevName, "write")
	}
	if readErr, ok := uintAt(stats, idxVDevReadErrors); ok {
		ch <- prometheus.MustNewConstMetric(c.vdevReadErrDesc, prometheus.CounterValue, float64(readErr), pool, vdevName)
	}
	if writeErr, ok := uintAt(stats, idxVDevWriteErrors); ok {
		ch <- prometheus.MustNewConstMetric(c.vdevWriteErrDesc, prometheus.CounterValue, float64(writeErr), pool, vdevName)
	}
	if cksumErr, ok := uintAt(stats, idxVDevChecksumErrors); ok {
		ch <- prometheus.MustNewConstMetric(c.vdevCksumErrDesc, prometheus.CounterValue, float64(cksumErr), pool, vdevName)
	}

	c.collectVdevChildren(ch, pool, vdevName, "data", vdev["children"], seen)
	c.collectVdevChildren(ch, pool, vdevName, "cache", vdev["l2cache"], seen)
	c.collectVdevChildren(ch, pool, vdevName, "spare", vdev["spares"], seen)
}

func (c *collector) collectVdevChildren(ch chan<- prometheus.Metric, pool, parent, vdevClass string, raw any, seen map[string]struct{}) {
	children, ok := sliceFromAny(raw)
	if !ok {
		return
	}
	for _, childRaw := range children {
		child, ok := mapFromAny(childRaw)
		if !ok {
			continue
		}
		c.collectVDev(ch, pool, parent, vdevClass, child, seen)
	}
}

func vdevLabel(vdev map[string]any, vdevType string) string {
	if p, ok := vdev["path"].(string); ok && len(p) != 0 {
		return p
	}
	if g, ok := uintFromAny(vdev["guid"]); ok {
		return fmt.Sprintf("%s:%d", vdevType, g)
	}
	if id, ok := uintFromAny(vdev["id"]); ok {
		return fmt.Sprintf("%s:%d", vdevType, id)
	}
	if len(vdevType) != 0 {
		return vdevType
	}
	return "unknown"
}

func uintAt(arr []uint64, idx int) (uint64, bool) {
	if idx < 0 || idx >= len(arr) {
		return 0, false
	}
	return arr[idx], true
}

func uintSliceFromAny(v any) []uint64 {
	switch vv := v.(type) {
	case []uint64:
		return vv
	case []any:
		out := make([]uint64, 0, len(vv))
		for _, item := range vv {
			u, ok := uintFromAny(item)
			if !ok {
				continue
			}
			out = append(out, u)
		}
		return out
	default:
		return nil
	}
}

func mapFromAny(v any) (map[string]any, bool) {
	out, ok := v.(map[string]any)
	return out, ok
}

func sliceFromAny(v any) ([]any, bool) {
	switch vv := v.(type) {
	case []any:
		return vv, true
	case []map[string]any:
		out := make([]any, 0, len(vv))
		for _, item := range vv {
			out = append(out, item)
		}
		return out, true
	default:
		return nil, false
	}
}

func uintFromAny(v any) (uint64, bool) {
	switch vv := v.(type) {
	case uint64:
		return vv, true
	case uint32:
		return uint64(vv), true
	case uint16:
		return uint64(vv), true
	case uint8:
		return uint64(vv), true
	case uint:
		return uint64(vv), true
	case int64:
		if vv < 0 {
			return 0, false
		}
		return uint64(vv), true
	case int32:
		if vv < 0 {
			return 0, false
		}
		return uint64(vv), true
	case int16:
		if vv < 0 {
			return 0, false
		}
		return uint64(vv), true
	case int8:
		if vv < 0 {
			return 0, false
		}
		return uint64(vv), true
	case int:
		if vv < 0 {
			return 0, false
		}
		return uint64(vv), true
	default:
		return 0, false
	}
}

func stateToString(state uint64) string {
	switch state {
	case uint64(ioctl.StateUnknown):
		return "unknown"
	case uint64(ioctl.StateClosed):
		return "closed"
	case uint64(ioctl.StateOffline):
		return "offline"
	case uint64(ioctl.StateRemoved):
		return "removed"
	case uint64(ioctl.StateCantOpen):
		return "cant_open"
	case uint64(ioctl.StateFaulted):
		return "faulted"
	case uint64(ioctl.StateDegraded):
		return "degraded"
	case uint64(ioctl.StateHealthy):
		return "healthy"
	default:
		return fmt.Sprintf("state_%d", state)
	}
}
