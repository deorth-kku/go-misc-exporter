//go:build ryzenadj

package main

import (
	"github.com/deorth-kku/go-misc-exporter/cmd/reg"
	"github.com/deorth-kku/go-misc-exporter/exporter/ryzenadj"
)

func init() {
	reg.Register("ryzenadj", ryzenadj.NewCollector)
}
