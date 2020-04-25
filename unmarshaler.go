package header

import (
	"errors"
	"net/http"
	"reflect"
	"strconv"
)

type Unmarshaler interface {
	UnmarshalHeader(string) error
}

func Unmarshal(headers http.Header, dst interface{}) error {
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("cannot parse")
	}

	rt := reflect.TypeOf(dst).Elem()
	for i := 0; i < rt.NumField(); i++ {
		fv := rv.Elem().Field(i)

		if fv.Kind() == reflect.Ptr && fv.IsNil() {
			fv.Set(reflect.New(fv.Type().Elem()))
		}

		if fv.Kind() != reflect.Ptr {
			fv = fv.Addr()
		}

		headerKey := rt.Field(i).Tag.Get("header")
		if p, ok := fv.Interface().(Unmarshaler); ok {
			if err := p.UnmarshalHeader(headers[headerKey][0]); err != nil {
				return err
			}
		} else if fv.Elem().Kind() == reflect.String {
			fv.Elem().Set(reflect.ValueOf(headers[headerKey][0]))
		} else if fv.Elem().Kind() == reflect.Int {
			val, err := strconv.Atoi(headers[headerKey][0])
			if err != nil {
				return err
			}
			fv.Elem().Set(reflect.ValueOf(val))
		}
	}

	return nil
}
