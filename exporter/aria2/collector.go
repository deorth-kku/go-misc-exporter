package aria2

import (
	"context"
	"log/slog"
	"slices"
	"strings"

	"github.com/deorth-kku/aria2rpc-go"
	ctime "github.com/deorth-kku/go-common/time"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	head        = "aria2_"
	global_head = head + "global_"
	task_head   = head + "task_"
)

type collector struct {
	servers          []*server
	path             string
	global_desc      map[string]*prometheus.Desc
	global_info_desc *prometheus.Desc
	task_desc        map[string]*prometheus.Desc
	task_info_desc   *prometheus.Desc
}

type server struct {
	*aria2rpc.Client
	ServerConf
}

func (s *server) connect() (err error) {
	if s.Client != nil {
		s.Client.Close()
	}
	s.Client, err = aria2rpc.New(context.Background(), s.Rpc, aria2rpc.WithSecret(s.Secret))
	return err
}

func NewCollector(conf Conf) (col *collector, err error) {
	col = &collector{
		path:        conf.Path,
		global_desc: make(map[string]*prometheus.Desc),
		task_desc:   make(map[string]*prometheus.Desc),
		servers:     make([]*server, len(conf.Servers)),
	}

	for i, serverConf := range conf.Servers {
		if serverConf.Timeout == 0 {
			serverConf.Timeout = 10
		}
		col.servers[i] = &server{
			ServerConf: serverConf,
		}
		err = col.servers[i].connect()
		if err != nil {
			return
		}
	}
	return
}

func (c *collector) Close() error {
	for _, conn := range c.servers {
		conn.Close()
	}
	return nil
}

var (
	server_label = []string{"server"}
	label_fields = []string{"bittorrent", "gid", "name"}
)

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	for k, v := range IterStructJson(aria2rpc.GlobalStat{}) {
		if _, ok := v.(uint); !ok {
			continue
		}
		c.global_desc[k] = prometheus.NewDesc(global_head+k, "", server_label, nil)
		ch <- c.global_desc[k]
	}
	task_info_labels := make([]string, 0)
	task_labels := append(server_label, label_fields...)
	for k, v := range IterStructJson(aria2rpc.Status{}) {
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
	c.global_info_desc = prometheus.NewDesc(global_head+"info", "", append(server_label, "version", "enabledFeatures"), nil)
	ch <- c.global_info_desc
}

func (c *collector) doServer(ch chan<- prometheus.Metric, server *server) {
	ctx, cancel := ctime.TimeoutContext(server.Timeout)
	defer cancel()
	gloabl_stat, err := server.GetGlobalStat(ctx)
	if err != nil {
		slog.Error("failed to get aria2 global stats", "server", server.Rpc, "err", err)
		return
	} else {
		for k, v := range IterStructJson(gloabl_stat) {
			vv, ok := v.(uint)
			if !ok {
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.global_desc[k], prometheus.GaugeValue, float64(vv), server.Rpc)
		}
	}
	version, err := server.GetVersion(ctx)
	if err != nil {
		slog.Error("failed to get aria2 global info", "server", server.Rpc, "err", err)
	} else {
		ch <- prometheus.MustNewConstMetric(c.global_info_desc, prometheus.GaugeValue, 0, server.Rpc, version.Version, strings.Join(version.EnabledFeatures, ","))
	}

	tasks, err := server.TellActive(ctx)
	if err != nil {
		slog.Error("failed to get aria2 tasks status", "server", server.Rpc, "err", err)
		return
	}

	for _, task := range tasks {
		labels := append([]string{server.Rpc}, TaskLabels(*task)...)
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

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	for _, server := range c.servers {
		c.doServer(ch, server)
	}
}

func (c *collector) Path() string {
	return c.path
}
