package cmd

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/deorth-kku/go-common"
)

type RawConf = map[string]json.RawMessage

var modules []string

func InitFlags() (rawconf RawConf, err error) {
	_, file, _, ok := runtime.Caller(1)
	if ok {
		modules = strings.Split(filepath.Base(filepath.Dir(file)), "+")
	} else {
		return nil, errors.New("failed to get modules")
	}

	var confpath string
	var help bool
	var install bool
	flag.StringVar(&confpath, "c", default_conf_file_path, "read config file")
	flag.BoolVar(&help, "h", false, "show help")
	flag.BoolVar(&install, "install", false, "install systemd service and default config file.")
	flag.Parse()
	if help {
		fmt.Println("go-misc-exporter built with modules:")
		for _, v := range modules {
			fmt.Printf("    %s\n", v)
		}
		fmt.Println()
		flag.Usage()
		os.Exit(0)
	}

	if install {
		err = install_service()
		if err != nil {
			return
		}
		os.Exit(0)
	}
	rawconf = make(RawConf)
	f, err := os.Open(confpath)
	if err != nil {
		return
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	err = dec.Decode(&rawconf)
	if err != nil {
		return
	}
	raw, ok := rawconf["exporter"]
	if ok {
		err = json.Unmarshal(raw, &conf)
		if err != nil {
			return
		}
	}
	if isRunningUnderSystemd() && len(conf.Log.File) == 0 {
		err = common.SetLog(conf.Log.File, conf.Log.Level, "DEFAULT", common.SlogHideTime{})
	} else {
		err = common.SetLog(conf.Log.File, conf.Log.Level, "TEXT")
	}
	return
}

func isRunningUnderSystemd() bool {
	_, exists := os.LookupEnv("INVOCATION_ID")
	return exists
}
