// Code generated by github.com/deorth-kku/go-misc-exporter/cmd. DO NOT EDIT.
package main

import (
	"encoding/json"
	"log/slog"

	"github.com/deorth-kku/go-misc-exporter/cmd"
	"github.com/deorth-kku/go-misc-exporter/exporter/nut"
	"github.com/deorth-kku/go-misc-exporter/exporter/smart"
)

func main() {
	rawconf, err := cmd.InitFlags()
	if err != nil {
		slog.Error("failed to init flags", "err", err)
		return
	}
	cs := make([]cmd.Collector, 0)

	nut_conf := nut.Conf{Path: cmd.DefaultMetricsPath}
	nut_raw, ok := rawconf["nut"]
	if ok {
		err = json.Unmarshal(nut_raw, &nut_conf)
		if err != nil {
			slog.Error("failed to parse conf", "section", "nut", "err", err)
			return
		}
		nut_col, err := nut.NewCollector(nut_conf)
		if err != nil {
			slog.Error("failed to init collector", "section", "nut", "err", err)
			return
		}
		cs = append(cs, nut_col)
	} else {
		slog.Info("setion not present, skipping", "exporter", "nut")
	}

	smart_conf := smart.Conf{Path: cmd.DefaultMetricsPath}
	smart_raw, ok := rawconf["smart"]
	if ok {
		err = json.Unmarshal(smart_raw, &smart_conf)
		if err != nil {
			slog.Error("failed to parse conf", "section", "smart", "err", err)
			return
		}
		smart_col, err := smart.NewCollector(smart_conf)
		if err != nil {
			slog.Error("failed to init collector", "section", "smart", "err", err)
			return
		}
		cs = append(cs, smart_col)
	} else {
		slog.Info("setion not present, skipping", "exporter", "smart")
	}

	err = cmd.StartCollectors(cs...)
	slog.Info("http server exited with", "err", err)
}
