package aria2

import (
	"iter"
	"reflect"
	"strings"
)

func IterStructJson(strc any) iter.Seq2[string, any] {
	v := reflect.ValueOf(strc)
	t := v.Type()
	return func(yield func(string, any) bool) {
		for i := range v.NumField() {
			jsontag, _, _ := strings.Cut(t.Field(i).Tag.Get("json"), ",")
			if !yield(jsontag, v.Field(i).Interface()) {
				return
			}
		}
	}
}
