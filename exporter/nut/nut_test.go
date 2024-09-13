package nut

import (
	"encoding/json"
	"testing"

	"github.com/deorth-kku/go-misc-exporter/cmd"
	"github.com/deorth-kku/go-misc-exporter/common"
)

var _ cmd.Collector = new(collector)

func TestCollector(t *testing.T) {
	var conf Conf
	err := json.Unmarshal(common.Must(cmd.InitFlags())["nut"], &conf)
	if err != nil {
		t.Error(err)
		return
	}

	col, err := NewCollector(conf)
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
