package cmd

import (
	"fmt"
	"io"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type Collector interface {
	prometheus.Collector
	io.Closer
	Path() string
}

func TestCollector(col Collector) (err error) {
	ch0 := make(chan *prometheus.Desc, 1)
	go func() {
		col.Describe(ch0)
		close(ch0)
	}()
	for desc, ok := <-ch0; ok; desc, ok = <-ch0 {
		fmt.Println(desc.String())
	}

	ch1 := make(chan prometheus.Metric, 1)
	go func() {
		col.Collect(ch1)
		close(ch1)
	}()
	dmetric := new(dto.Metric)
	for metr, ok := <-ch1; ok; metr, ok = <-ch1 {
		err = metr.Write(dmetric)
		if err != nil {
			col.Close()
			return
		}
		fmt.Println(dmetric.String())
	}
	return
}

func TestCollectorThenClose(col Collector) (err error) {
	defer col.Close()
	return TestCollector(col)
}
