package nut

import (
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"github.com/deorth-kku/go-misc-exporter/common"
	"github.com/prometheus/client_golang/prometheus"
	nut "github.com/robbiet480/go.nut"
)

const head = "nut_"

var (
	labels_template = []string{"server", "ups"}
)

type collector struct {
	clients         []nut.Client
	server_info     *prometheus.Desc
	ups_info        *prometheus.Desc
	ups_info_labels []string
	descs           map[string]*prometheus.Desc
	info_labels     []string
	path            string
}

func NewCollector(conf Conf) (c *collector, err error) {
	c = new(collector)
	c.path = conf.Path
	c.descs = make(map[string]*prometheus.Desc)
	c.clients = make([]nut.Client, len(conf.Servers))
	for i, server := range conf.Servers {
		c.clients[i], err = Connect(server.Host, server.Port, common.FloatDuration(server.Timeout))
		if err != nil {
			return nil, err
		}
		if len(server.Username)+len(server.Password) != 0 {
			_, err = c.clients[i].Authenticate(server.Username, server.Password)
			if err != nil {
				slog.Error("Authenticate failed", "server", server, "err", err)
				return nil, err
			}
		}
	}
	return
}

func descname(value_name string) string {
	return head + strings.ReplaceAll(value_name, ".", "_")
}

func (c *collector) Close() (err error) {
	for _, client := range c.clients {
		_, err = client.Disconnect()
		if err != nil {
			slog.Error("disconnet failed", "server", client.Hostname.String(), "err", err)
		}
	}
	return
}

func (c *collector) Path() string {
	return c.path
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	var ok bool
	c.server_info = prometheus.NewDesc(head+"server_info", "nut server information", []string{"server", "Version", "ProtocolVersion"}, nil)
	for _, client := range c.clients {
		for _, ups := range common.Must(client.GetUPSList()) {
			for _, v := range ups.Variables {
				if _, ok = c.descs[v.Name]; ok {
					continue
				}
				switch v.Value.(type) {
				case int64, float64:
					c.descs[v.Name] = prometheus.NewDesc(descname(v.Name), v.Description, labels_template, nil)
				case string, bool:
					if !slices.Contains(c.ups_info_labels, v.Name) {
						c.ups_info_labels = append(c.ups_info_labels, v.Name)
					}
				default:
					slog.Warn("unexpected variable type", "name", v.Name, "type", v.Type)
				}
			}
		}
	}

	c.ups_info = prometheus.NewDesc(head+"ups_info", "nut ups information, value is for number of logins", append(append(labels_template, "description", "master"), fix_label(c.ups_info_labels)...), nil)
	ch <- c.server_info
	ch <- c.ups_info
	for _, desc := range c.descs {
		ch <- desc
	}
}

func fix_label(in []string) (out []string) {
	out = make([]string, len(in))
	for i, v := range in {
		out[i] = strings.ReplaceAll(v, ".", "_")
	}
	return
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	for _, client := range c.clients {
		serverstring := client.Hostname.String()
		ch <- prometheus.MustNewConstMetric(c.server_info, prometheus.GaugeValue, 0, serverstring, client.Version, client.ProtocolVersion)
		for _, ups := range common.Must(client.GetUPSList()) {
			info_map := make(map[string]string)
			for _, v := range ups.Variables {
				desc, ok := c.descs[v.Name]
				if !ok && !slices.Contains(c.ups_info_labels, v.Name) {
					continue
				}
				switch vv := v.Value.(type) {
				case int64:
					ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, float64(vv), serverstring, ups.Name)
				case float64:
					ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, vv, serverstring, ups.Name)
				case string:
					info_map[v.Name] = vv
				case bool:
					info_map[v.Name] = strconv.FormatBool(vv)
				}
			}
			info_list := make([]string, len(c.ups_info_labels)+4)
			info_list[0] = serverstring
			info_list[1] = ups.Name
			info_list[2] = ups.Description
			info_list[3] = strconv.FormatBool(ups.Master)
			for i, label := range c.ups_info_labels {
				info_list[i+4] = info_map[label]
			}
			ch <- prometheus.MustNewConstMetric(c.ups_info, prometheus.GaugeValue, float64(ups.NumberOfLogins), info_list...)
		}
	}
	return
}
