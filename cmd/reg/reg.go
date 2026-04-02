package reg

import (
	"encoding/json"
	"reflect"

	"github.com/deorth-kku/go-common"
	"github.com/deorth-kku/go-misc-exporter/cmd"
)

type regitem struct {
	fn func(any) (cmd.Collector, error)
	ty reflect.Type
}

var regs = make(map[string]regitem)

func Register[CONF any, COL cmd.Collector](name string, fn func(CONF) (COL, error)) {
	regs[name] = regitem{
		fn: func(a any) (cmd.Collector, error) {
			return fn(a.(CONF))
		},
		ty: reflect.TypeFor[CONF](),
	}
}

const ErrNotRegisted = common.ErrorString("no such collector")

func NewCollector(name string, data json.RawMessage) (cmd.Collector, error) {
	item, ok := regs[name]
	if !ok {
		return nil, ErrNotRegisted
	}
	conf := reflect.New(item.ty)
	field := conf.Elem().FieldByName("Path")
	if field.IsValid() && field.CanSet() && field.Kind() == reflect.String {
		field.SetString(cmd.DefaultMetricsPath)
	}
	err := json.Unmarshal(data, conf.Interface())
	if err != nil {
		return nil, err
	}
	return item.fn(conf.Elem().Interface())
}
