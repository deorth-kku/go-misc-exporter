// Code generated by github.com/deorth-kku/go-misc-exporter/cmd. DO NOT EDIT.
package main

import (
	"encoding/json"
	"log/slog"

	"github.com/deorth-kku/go-misc-exporter/cmd"
	"github.com/deorth-kku/go-misc-exporter/exporter/aria2"
	"github.com/deorth-kku/go-misc-exporter/exporter/ryzenadj"
)

func main() {
	rawconf, err := cmd.InitFlags()
	if err != nil {
		slog.Error("failed to init flags", "err", err)
		return
	}
	cs := make([]cmd.Collector, 0)

	aria2_conf := aria2.Conf{Path: "/metrics"}
	aria2_raw, ok := rawconf["aria2"]
	if ok {
		err = json.Unmarshal(aria2_raw, &aria2_conf)
		if err != nil {
			slog.Error("failed to parse conf", "section", "aria2", "err", err)
			return
		}
		aria2_col, err := aria2.NewCollector(aria2_conf)
		if err != nil {
			slog.Error("failed to init collector", "section", "aria2", "err", err)
		}
		cs = append(cs, aria2_col)
	} else {
		slog.Info("setion not present, skipping", "exporter", "aria2")
	}

	ryzenadj_conf := ryzenadj.Conf{Path: "/metrics"}
	ryzenadj_raw, ok := rawconf["ryzenadj"]
	if ok {
		err = json.Unmarshal(ryzenadj_raw, &ryzenadj_conf)
		if err != nil {
			slog.Error("failed to parse conf", "section", "ryzenadj", "err", err)
			return
		}
		ryzenadj_col, err := ryzenadj.NewCollector(ryzenadj_conf)
		if err != nil {
			slog.Error("failed to init collector", "section", "ryzenadj", "err", err)
		}
		cs = append(cs, ryzenadj_col)
	} else {
		slog.Info("setion not present, skipping", "exporter", "ryzenadj")
	}

	err = cmd.StartCollectors(cs...)
	slog.Info("http server exited with", "err", err)
}
