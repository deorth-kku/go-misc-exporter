package nut

import (
	"testing"

	"github.com/deorth-kku/go-misc-exporter/cmd"
)

var _ cmd.Collector = new(collector)

func TestCollector(t *testing.T) {
	col, err := NewCollector(Conf{
		Servers: []Server{{
			Host:     "localhost",
			Username: "admin",
			Password: "mypass",
		}},
	})
	if err != nil {
		t.Error(err)
		return
	}
	err = cmd.TestCollectorThenClose(col)
	if err != nil {
		t.Error(err)
		return
	}
}
