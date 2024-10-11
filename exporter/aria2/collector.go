package aria2

import (
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/deorth-kku/go-common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/siku2/arigo"
)

const (
	head        = "aria2_"
	global_head = head + "global_"
	task_head   = head + "task_"
)

type collector struct {
	servers        []*server
	path           string
	global_desc    map[string]*prometheus.Desc
	task_desc      map[string]*prometheus.Desc
	task_info_desc *prometheus.Desc
}

type server struct {
	*arigo.Client
	ServerConf
}

func (s *server) connect() (err error) {
	ctx, cancal := common.TimeoutContext(s.Timeout)
	defer cancal()
	if s.Client != nil {
		s.Client.Close()
	}
	s.Client, err = arigo.DialContext(ctx, s.Rpc, s.Secret)
	return err
}

const (
	reconnect_times    = 3
	reconnect_interval = 1 * time.Second
)

func (s *server) reconnect() error {
	var err error
	bak := s.Client
	slog.Warn("try to reconnect to aria2 rpc", "rpc", s.Rpc)
	for i := range reconnect_times {
		err = s.connect()
		if err != nil {
			slog.Warn("failed to reconnect aria2 rpc", "rpc", s.Rpc, "retry", fmt.Sprintf("%d/%d", i+1, reconnect_times), "err", err)
			time.Sleep(reconnect_interval)
		} else {
			return nil
		}
	}
	s.Client = bak
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
			err = col.servers[i].reconnect()
			if err != nil {
				return
			}
		}
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
		gloabl_stat, err := server.GetGlobalStats()
		if err != nil {
			slog.Error("failed to get aria2 global stats", "server", server.Rpc, "err", err)
			server.reconnect()
		} else {
			for k, v := range IterStructJson(gloabl_stat) {
				vv, ok := v.(uint)
				if !ok {
					continue
				}
				ch <- prometheus.MustNewConstMetric(c.global_desc[k], prometheus.GaugeValue, float64(vv), server.Rpc)
			}
		}

		tasks, err := server.TellActive([]string{}...)
		if err != nil {
			slog.Error("failed to get aria2 tasks status", "server", server.Rpc, "err", err)
			server.reconnect()
			return
		}

		for _, task := range tasks {
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
