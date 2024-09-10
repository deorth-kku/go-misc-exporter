package smart

import (
	"github.com/anatol/smart.go"
	"github.com/prometheus/client_golang/prometheus"
)

// since smart.go does not fully support scsi, only info is provided. All metrics are not available

type ScsiDev struct {
	name string
	dev  *smart.ScsiDevice
}

const (
	metric_scsi = metric_head + "scsi_"
	scsi_info   = metric_scsi + "Info"
)

func (d *ScsiDev) ListMetrics() map[string]*prometheus.Desc {
	d.dev.Inquiry()
	return nil
}
