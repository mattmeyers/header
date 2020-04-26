package header

import (
	"net/http"
	"reflect"
	"strconv"
)

type Marshaler interface {
	MarshalHeader() (string, error)
}

type Stringer interface {
	String() string
}

func Marshal(v interface{}) (http.Header, error) {
	h := make(http.Header)
	rv := reflect.ValueOf(v)
	rt := reflect.TypeOf(v)
	for i := 0; i < rt.NumField(); i++ {
		fv := rv.Field(i)
		ft := rt.Field(i).Tag.Get("header")
		if ft == "" {
			continue
		}

		if m, ok := fv.Interface().(Marshaler); ok {
			s, err := m.MarshalHeader()
			if err != nil {
				return nil, err
			} else if s == "" {
				continue
			}
			h.Add(ft, s)
		} else if m, ok := fv.Interface().(Stringer); ok {
			s := m.String()
			if s == "" {
				continue
			}
			h.Add(ft, s)
		} else if fv.Kind() == reflect.String {
			h.Add(ft, fv.String())
		} else if fv.Kind() == reflect.Int {
			h.Add(ft, strconv.Itoa(int(fv.Int())))
		}
	}

	return h, nil
}
