// Code generated by github.com/deorth-kku/go-misc-exporter/cmd. DO NOT EDIT.
package main

import (
	"encoding/json"
	"log/slog"

	"github.com/deorth-kku/go-misc-exporter/cmd"
	"github.com/deorth-kku/go-misc-exporter/exporter/hwmon"
	"github.com/deorth-kku/go-misc-exporter/exporter/nut"
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

	hwmon_conf := hwmon.Conf{Path: "/metrics"}
	hwmon_raw, ok := rawconf["hwmon"]
	if ok {
		err = json.Unmarshal(hwmon_raw, &hwmon_conf)
		if err != nil {
			slog.Error("failed to parse conf", "section", "hwmon", "err", err)
			return
		}
		hwmon_col, err := hwmon.NewCollector(hwmon_conf)
		if err != nil {
			slog.Error("failed to init collector", "section", "hwmon", "err", err)
		}
		cs = append(cs, hwmon_col)
	} else {
		slog.Info("setion not present, skipping", "exporter", "hwmon")
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
