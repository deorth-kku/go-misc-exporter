//go:build smart

package main

import (
	"github.com/deorth-kku/go-misc-exporter/cmd/reg"
	"github.com/deorth-kku/go-misc-exporter/exporter/smart"
)

func init() {
	reg.Register("smart", smart.NewCollector)
}
