package status

import (
	"fmt"
	"reflect"

	"github.com/fatih/structs"
)

// Status converts stats into pct map status
type Status struct {
	stats interface{}
}

func New(stats interface{}) *Status {
	// initialize pointers for all struct fields
	v := reflect.Indirect(reflect.ValueOf(stats))
	for i := 0; i < v.NumField(); i++ {
		v.Field(i).Set(reflect.New(v.Field(i).Type().Elem()))
	}

	return &Status{
		stats: stats,
	}
}

// Map converts stats struct into a map
func (s *Status) Map() map[string]string {
	out := map[string]string{}
	for _, f := range structs.New(s.stats).Fields() {
		if f.IsZero() {
			continue
		}
		tag := f.Tag("name")
		if tag == "" {
			continue
		}
		v := fmt.Sprint(f.Value())
		if v == "" || v == `""` || v == "0" {
			continue
		}
		out[tag] = v
	}
	return out
}
