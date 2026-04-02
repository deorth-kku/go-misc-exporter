package main

import (
	"log/slog"

	"github.com/deorth-kku/go-misc-exporter/cmd"
	"github.com/deorth-kku/go-misc-exporter/cmd/reg"
)

func main() {
	rawconf, err := cmd.InitFlags()
	if err != nil {
		slog.Error("failed to init flags", "err", err)
		return
	}
	cs := make([]cmd.Collector, 0, len(rawconf))
	for name, data := range rawconf {
		c, err := reg.NewCollector(name, data)
		if err != nil {
			slog.Error("failed to init collector", "section", name, "err", err)
			return
		}
		cs = append(cs, c)
	}
	err = cmd.StartCollectors(cs...)
	slog.Info("http server exited with", "err", err)
}
