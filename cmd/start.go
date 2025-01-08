package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	rtpprof "runtime/pprof"
	"slices"
	"strings"
	"sync"
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
	Pprof  PprofPath   `json:"pprof,omitempty"`
}

type PprofPath map[string]string

type handlerfunc = func(http.ResponseWriter, *http.Request)

func (pp PprofPath) Handlers(yield func(string, handlerfunc) bool) {
	var f handlerfunc
	for name, path := range pp {
		switch name {
		case "profile":
			f = pprof.Profile
		case "cmdline":
			f = pprof.Cmdline
		case "trace":
			f = pprof.Trace
		case "symbol":
			f = pprof.Symbol
		default:
			if rtpprof.Lookup(name) == nil {
				continue
			}
			f = pprof.Handler(name).ServeHTTP
		}
		if !yield(path, f) {
			return
		}
	}
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
	for path, hfunc := range conf.Pprof.Handlers {
		server.HandleFunc(path, hfunc)
	}
	var wg sync.WaitGroup
	var close_errs []any
	wg.Add(1)
	common.SignalsCallback(func() {
		server.Shutdown(context.Background())
		for _, c := range cs {
			err := c.Close()
			if err != nil {
				close_errs = append(close_errs, err)
			}
		}
		wg.Done()
	}, true, syscall.SIGINT, syscall.SIGTERM)
	err = server.ListenAndServe(conf.Listen)
	wg.Wait()
	if len(close_errs) != 0 {
		errfmt := "http server exit: %w, collecter close errors: %w" + strings.Repeat(", %w", len(close_errs)-1)
		close_errs = slices.Insert(close_errs, 0, any(err))
		err = fmt.Errorf(errfmt, close_errs...)
	}
	return
}
