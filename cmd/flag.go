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
	}
	return
}
