package cmd

import (
	"context"
	"log/slog"
	"syscall"

	"github.com/deorth-kku/go-common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type LogSettings struct {
	File  string `json:"file,omitempty"`
	Level string `json:"level,omitempty"`
}

type Conf struct {
	Listen string      `json:"listen,omitempty"`
	Path   string      `json:"path,omitempty"`
	Log    LogSettings `json:"log,omitempty"`
}

var conf = Conf{
	Log: LogSettings{
		Level: "INFO",
	},
}

func StartCollectors(cs ...Collector) (err error) {
	paths := make(map[string][]prometheus.Collector)
	for _, c := range cs {
		paths[c.Path()] = append(paths[c.Path()], c)
	}

	server := common.NewHttpServer()

	for path, cs := range paths {
		r := prometheus.NewRegistry()
		r.MustRegister(cs...)
		handler := promhttp.HandlerFor(r, promhttp.HandlerOpts{})
		server.Handle(path, handler)
	}

	if len(conf.Path) != 0 {
		server.Handle(conf.Path, promhttp.Handler())

	}
	common.SignalsCallback(func() { server.Shutdown(context.Background()) }, true, syscall.SIGINT, syscall.SIGTERM)
	err = server.ListenAndServe(conf.Listen)
	if err != nil {
		slog.Error("http server exit with error", "err", err)
	}
	return
}
