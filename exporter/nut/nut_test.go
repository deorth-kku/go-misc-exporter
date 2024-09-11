package nut

import (
	"net/http"
	"testing"

	"github.com/deorth-kku/go-misc-exporter/cmd"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var _ cmd.Collector = new(collector)

func TestCollector(t *testing.T) {
	col, _ := NewCollector(Conf{
		Servers: []Server{{
			Host:     "localhost",
			Username: "admin",
			Password: "mypass",
		}},
	})
	prometheus.MustRegister(col)
	http.ListenAndServe(":8188", promhttp.Handler())
}
