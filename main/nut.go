//go:build nut

package main

import (
	"github.com/deorth-kku/go-misc-exporter/cmd/reg"
	"github.com/deorth-kku/go-misc-exporter/exporter/nut"
)

func init() {
	reg.Register("nut", nut.NewCollector)
}
