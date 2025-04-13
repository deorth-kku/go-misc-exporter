package smart

import (
	"github.com/anatol/smart.go"
	"github.com/deorth-kku/go-common"
	"github.com/prometheus/client_golang/prometheus"
)

type PromDev interface {
	Name() string
	ListMetrics() map[string]*prometheus.Desc
	GetMetrics() []PromValue
	Close() error
}

type PromValue struct {
	Desc  *prometheus.Desc
	Type  prometheus.ValueType
	Value float64
	Tags  []string
}

func NewPromDev(name string) (d PromDev, err error) {
	dev, err := smart.Open("/dev/" + name)
	if err != nil {
		return
	}
	switch sm := dev.(type) {
	case *smart.SataDevice:
		d = NewSataDev(name, sm)
	case *smart.NVMeDevice:
		d = NewNvmeDev(name, sm)
	case *smart.ScsiDevice:
		return nil, common.ErrorString("no support for scsi devices yet")
	default:
		return nil, common.ErrorString("unknown device type")
	}
	return
}
