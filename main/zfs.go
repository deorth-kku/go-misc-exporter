//go:build zfs

package main

import (
	"github.com/deorth-kku/go-misc-exporter/cmd/reg"
	"github.com/deorth-kku/go-misc-exporter/exporter/zfs"
)

func init() {
	reg.Register("zfs", zfs.NewCollector)
}
