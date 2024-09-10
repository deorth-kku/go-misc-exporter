package cmd

import (
	"encoding/json"
	"flag"
	"os"
)

type RawConf = map[string]json.RawMessage

func InitFlags() (rawconf RawConf, err error) {
	var confpath string
	var help bool
	flag.StringVar(&confpath, "c", "/etc/gme/conf.json", "read config file")
	flag.BoolVar(&help, "h", false, "show help")
	flag.Parse()
	if help {
		flag.Usage()
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
