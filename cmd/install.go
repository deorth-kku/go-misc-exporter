package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/coreos/go-systemd/v22/unit"
	"github.com/deorth-kku/go-common"
)

const (
	service_file_path       = "/etc/systemd/system/go-misc-exporter.service"
	alias_service_file_path = "/etc/systemd/system/gme.service"
	DefaultMetricsPath      = "/metrics"
)

var default_conf_file_path = func() string {
	if runtime.GOOS == "windows" {
		return ".\\conf.json"
	}
	return "/etc/gme/conf.json"
}()

const ErrUserCanceled = common.ErrorString("user canceled")

func install_service() (err error) {
	if runtime.GOOS != "linux" {
		return errors.New("only support linux systemd service install")
	}
	conf := make(RawConf)
	_, err = os.Stat(default_conf_file_path)
	if os.IsNotExist(err) {
		slog.Info("config file does not exist, creating", "file", default_conf_file_path)
		var data []byte
		data, err = json.Marshal(Conf{Listen: ":4403"})
		if err != nil {
			return
		}
		conf["exporter"] = data
		err = os.Mkdir(filepath.Dir(default_conf_file_path), 0755)
		if errors.Is(err, os.ErrExist) {
		} else if err != nil {
			return
		}
		f, err := os.Create(default_conf_file_path)
		if err != nil {
			return err
		}
		defer f.Close()
		enc := json.NewEncoder(f)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "  ")
		err = enc.Encode(conf)
		if err != nil {
			return err
		}
	} else if err == nil {
		slog.Info("config file exists, verify", "conf", default_conf_file_path)
		f, err := os.Open(default_conf_file_path)
		if err != nil {
			return err
		}
		defer f.Close()
		decoder := json.NewDecoder(f)
		err = decoder.Decode(&conf)
		if err != nil {
			return err
		}
		slog.Info("verify ok")
	} else {
		return
	}

	_, err = os.Stat(service_file_path)
	if os.IsNotExist(err) {
		err = create_service_file(conf)
	} else if err == nil {
		var ok bool
		ok, err = askQuestionBool("File existed, do you want to overwrite %s?", service_file_path)
		if err != nil {
			return
		}
		if ok {
			err = create_service_file(conf)
		}
	} else {
		return
	}
	if err != nil {
		return
	}

	err = os.Symlink(service_file_path, alias_service_file_path)
	if os.IsExist(err) {
		slog.Info("skip creating alias, because file already exist", "file", alias_service_file_path)
		err = nil
	}
	if err == nil {
		slog.Info("you might want to run \"systemctl daemon-reload\"")
	}
	return
}

func create_service_file(conf RawConf) (err error) {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	setion_unit := &unit.UnitSection{
		Section: "Unit",
		Entries: []*unit.UnitEntry{
			{
				Name:  "Description",
				Value: "Exporter for some services",
			},
			{
				Name:  "Documentation",
				Value: "https://github.com/deorth-kku/go-misc-exporter",
			},
			{
				Name:  "After",
				Value: "network.target",
			},
		},
	}

	var req []string
	_, ok := conf["nut"]
	if slices.Contains(modules, "nut") && ok {
		req = append(req, "nut-server.service")
	}
	_, ok = conf["aria2"]
	if slices.Contains(modules, "aria2") && ok {
		req = append(req, "aria2.service")
	}
	if len(req) != 0 {
		setion_unit.Entries = append(setion_unit.Entries, &unit.UnitEntry{
			Name:  "Requires",
			Value: strings.Join(req, " "),
		})
	}

	setion_service := &unit.UnitSection{
		Section: "Service",
		Entries: []*unit.UnitEntry{
			{
				Name:  "User",
				Value: "root",
			},
			{
				Name:  "ExecStart",
				Value: exe,
			},
			{
				Name:  "Restart",
				Value: "on-failure",
			},
		},
	}

	setion_install := &unit.UnitSection{
		Section: "Install",
		Entries: []*unit.UnitEntry{
			{
				Name:  "Alias",
				Value: "gme.service",
			},
			{
				Name:  "WantedBy",
				Value: "multi-user.target",
			},
		},
	}
	reader := unit.SerializeSections([]*unit.UnitSection{setion_unit, setion_service, setion_install})
	f, err := os.Create(service_file_path)
	if err != nil {
		return
	}
	defer f.Close()
	_, err = io.Copy(f, reader)
	return
}

func askQuestionBool(q string, args ...any) (a bool, err error) {
	data := make([]byte, 1)

	for {
		fmt.Printf(q+" (y/N)\n", args...)
		_, err = os.Stdin.Read(data)
		if err != nil {
			return
		}
		switch data[0] {
		case 'y', 'Y':
			a = true
			return
		case 'n', 'N', '\n':
			a = false
			return
		}
	}
}
