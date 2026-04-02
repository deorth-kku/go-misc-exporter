//go:build systemd

package main

import (
	"github.com/deorth-kku/go-misc-exporter/cmd/reg"
	"github.com/deorth-kku/go-misc-exporter/exporter/systemd"
)

func init() {
	reg.Register("systemd", systemd.NewCollector)
}
