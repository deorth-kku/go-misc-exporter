package systemd

import (
	"fmt"
	"maps"
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/deorth-kku/go-misc-exporter/cmd"
	"github.com/deorth-kku/go-misc-exporter/common"
)

func TestGenProps(t *testing.T) {
	ctx, cancel := common.TimeoutContext(10.0)
	defer cancel()
	d, err := dbus.NewSystemdConnectionContext(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	defer d.Close()
	units, err := d.ListUnitsContext(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	types := make(map[string]string)
	for _, unit := range units {
		prop, err := d.GetAllPropertiesContext(ctx, unit.Name)
		if err != nil {
			t.Error(err)
			return
		}
		for k, v := range prop {
			types[k] = reflect.TypeOf(v).String()
		}
	}
	f, err := os.Create("./allprops.txt")
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()
	for _, k := range slices.Sorted(maps.Keys(types)) {
		fmt.Fprintf(f, "%s: %s\n", k, types[k])
	}
}

func TestCollector(t *testing.T) {
	err := cmd.TestCollectorThenClose(common.Must(NewCollector(Conf{
		States:   []string{"active"},
		Patterns: []string{"*.service"},
	})))
	if err != nil {
		t.Error(err)
		return
	}
}

func TestParseTime(t *testing.T) {
	t0, err := parseTimeStamp("@t 1722171246670483")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(t0)
}
