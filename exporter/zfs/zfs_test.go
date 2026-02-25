package zfs

import (
	"os"
	"testing"

	"github.com/deorth-kku/go-misc-exporter/cmd"
)

var _ cmd.Collector = new(collector)

func TestMatchSkip(t *testing.T) {
	c := Conf{Skip: []string{"tank"}}
	if !c.MatchSkip("tank") {
		t.Error("failed to match skipped pool")
	}
	if c.MatchSkip("data") {
		t.Error("unexpected skip match")
	}
}

func TestCollector(t *testing.T) {
	if _, err := os.Stat("/dev/zfs"); err != nil {
		t.Skipf("skip zfs collector test: /dev/zfs not available: %v", err)
	}

	col, err := NewCollector(Conf{Path: cmd.DefaultMetricsPath})
	if err != nil {
		t.Skipf("skip zfs collector test: %v", err)
	}

	if err := cmd.TestCollectorThenClose(col); err != nil {
		t.Error(err)
	}
}
