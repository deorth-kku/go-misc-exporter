//go:build aria2

package main

import (
	"github.com/deorth-kku/go-misc-exporter/cmd/reg"
	"github.com/deorth-kku/go-misc-exporter/exporter/aria2"
)

func init() {
	reg.Register("aria2", aria2.NewCollector)
}
