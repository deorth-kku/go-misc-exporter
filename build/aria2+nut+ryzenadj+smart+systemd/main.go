// Code generated by github.com/deorth-kku/go-misc-exporter/cmd. DO NOT EDIT.
package main

import (
	"encoding/json"
	"log/slog"

	"github.com/deorth-kku/go-misc-exporter/cmd"
	"github.com/deorth-kku/go-misc-exporter/exporter/aria2"
	"github.com/deorth-kku/go-misc-exporter/exporter/nut"
	"github.com/deorth-kku/go-misc-exporter/exporter/ryzenadj"
	"github.com/deorth-kku/go-misc-exporter/exporter/smart"
	"github.com/deorth-kku/go-misc-exporter/exporter/systemd"
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

	nut_conf := nut.Conf{Path: "/metrics"}
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
		}
		cs = append(cs, nut_col)
	} else {
		slog.Info("setion not present, skipping", "exporter", "nut")
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

	smart_conf := smart.Conf{Path: "/metrics"}
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
		}
		cs = append(cs, smart_col)
	} else {
		slog.Info("setion not present, skipping", "exporter", "smart")
	}

	systemd_conf := systemd.Conf{Path: "/metrics"}
	systemd_raw, ok := rawconf["systemd"]
	if ok {
		err = json.Unmarshal(systemd_raw, &systemd_conf)
		if err != nil {
			slog.Error("failed to parse conf", "section", "systemd", "err", err)
			return
		}
		systemd_col, err := systemd.NewCollector(systemd_conf)
		if err != nil {
			slog.Error("failed to init collector", "section", "systemd", "err", err)
		}
		cs = append(cs, systemd_col)
	} else {
		slog.Info("setion not present, skipping", "exporter", "systemd")
	}

	err = cmd.StartCollectors(cs...)
	slog.Info("http server exited with", "err", err)
}
