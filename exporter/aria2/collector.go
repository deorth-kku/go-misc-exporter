package aria2

import (
	"context"
	"slices"

	"github.com/deorth-kku/go-misc-exporter/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/siku2/arigo"
)

const (
	head        = "aria2_"
	global_head = head + "global_"
	task_head   = head + "task_"
)

type collector struct {
	servers        []server
	path           string
	global_desc    map[string]*prometheus.Desc
	task_desc      map[string]*prometheus.Desc
	task_info_desc *prometheus.Desc
}

type server struct {
	*arigo.Client
	Rpc string
}

func NewCollector(conf Conf) (col *collector, err error) {
	col = &collector{
		path:        conf.Path,
		global_desc: make(map[string]*prometheus.Desc),
		task_desc:   make(map[string]*prometheus.Desc),
		servers:     make([]server, len(conf.Servers)),
	}

	for i, server := range conf.Servers {
		if server.Timeout == 0 {
			server.Timeout = 10
		}
		ctx, cancal := context.WithTimeout(context.Background(), common.FloatDuration(server.Timeout))
		defer cancal()
		col.servers[i].Client, err = arigo.DialContext(ctx, server.Rpc, server.Secret)
		if err != nil {
			return
		}
		col.servers[i].Rpc = server.Rpc
	}
	return
}

func (c *collector) Close() (err error) {
	for _, conn := range c.servers {
		e := conn.Close()
		if e != nil {
			err = e
		}
	}
	return
}

var (
	server_label = []string{"server"}
	label_fields = []string{"bittorrent", "gid", "name"}
)

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	for k, v := range IterStructJson(arigo.Stats{}) {
		if _, ok := v.(uint); !ok {
			continue
		}
		c.global_desc[k] = prometheus.NewDesc(global_head+k, "", server_label, nil)
		ch <- c.global_desc[k]
	}
	task_info_labels := make([]string, 0)
	task_labels := append(server_label, label_fields...)
	for k, v := range IterStructJson(arigo.Status{}) {
		if slices.Contains(label_fields, k) {
			continue
		}
		switch v.(type) {
		case uint:
			c.task_desc[k] = prometheus.NewDesc(task_head+k, "", task_labels, nil)
			ch <- c.task_desc[k]
		case string:
			task_info_labels = append(task_info_labels, k)
		}
	}
	c.task_info_desc = prometheus.NewDesc(task_head+"info", "", append(task_labels, task_info_labels...), nil)
	ch <- c.task_info_desc
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	for _, server := range c.servers {
		for k, v := range IterStructJson(common.Must(server.GetGlobalStats())) {
			vv, ok := v.(uint)
			if !ok {
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.global_desc[k], prometheus.GaugeValue, float64(vv), server.Rpc)
		}

		for _, task := range common.Must(server.TellActive([]string{}...)) {
			labels := append([]string{server.Rpc}, TaskLabels(task)...)
			info_labels := slices.Clone(labels)
			for k, v := range IterStructJson(task) {
				if slices.Contains(label_fields, k) {
					continue
				}
				switch vv := v.(type) {
				case uint:
					ch <- prometheus.MustNewConstMetric(c.task_desc[k], prometheus.GaugeValue, float64(vv), labels...)
				case string:
					info_labels = append(info_labels, vv)
				}
			}
			ch <- prometheus.MustNewConstMetric(c.task_info_desc, prometheus.GaugeValue, 0, info_labels...)
		}
	}
}

func (c *collector) Path() string {
	return c.path
}
