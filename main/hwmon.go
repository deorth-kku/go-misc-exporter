//go:build hwmon

package main

import (
	"github.com/deorth-kku/go-misc-exporter/cmd/reg"
	"github.com/deorth-kku/go-misc-exporter/exporter/hwmon"
)

func init() {
	reg.Register("hwmon", hwmon.NewCollector)
}
