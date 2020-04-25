package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type Unmarshaler interface {
	Unmarshal(string) error
}

type StringHeader struct {
	Value string
}

func (h *StringHeader) Parse(s string) error {
	h.Value = s
	return nil
}

type Server struct{ StringHeader }

func NewServer(s string) *Server {
	h := &Server{}
	h.Parse(s)
	return h
}

type ContentType struct {
	MediaType string
	Charset   string
	Boundary  string
}

func NewContentType(s string) (*ContentType, error) {
	h := &ContentType{}
	err := h.Unmarshal(s)
	if err != nil {
		return nil, err
	}

	return h, nil
}

func (h *ContentType) Unmarshal(s string) error {
	parts := strings.Split(s, ";")
	if len(parts) == 0 {
		return nil
	}

	mTypes := strings.Split(parts[0], "/")
	if len(mTypes) != 2 || len(mTypes[0]) == 0 || len(mTypes[1]) == 0 {
		return errors.New("invalid media type")
	}
	h.MediaType = parts[0]

	for _, p := range parts[1:] {
		a := strings.Split(p, "=")
		if len(a) != 2 {
			return errors.New("malformed Content-Type header")
		}
		switch strings.TrimSpace(a[0]) {
		case "charset":
			h.Charset = a[1]
		case "boundary":
			h.Boundary = a[1]
		default:
			return errors.New("invalid Content-Type directive")
		}
	}

	return nil
}

// ContentDisposition holds the data held in a Content-Disposition header.
//
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Disposition
type ContentDisposition struct {
	Type      string
	Name      string
	Filename  string
	FilenameS string
}

func NewContentDisposition(s string) (*ContentDisposition, error) {
	h := &ContentDisposition{}
	err := h.Unmarshal(s)
	if err != nil {
		return nil, err
	}

	return h, nil
}

func (h *ContentDisposition) Unmarshal(s string) error {
	parts := strings.Split(s, ";")
	if len(parts) == 0 {
		return nil
	}

	if parts[0] != "inline" && parts[0] != "attachment" && parts[0] != "form-data" {
		return errors.New("invalid Content-Disposition type")
	}
	h.Type = parts[0]

	for _, p := range parts[1:] {
		a := strings.Split(p, "=")
		if len(a) != 2 {
			return errors.New("malformed Content-Disposition header")
		}
		switch strings.TrimSpace(a[0]) {
		case "name":
			h.Name = strings.Trim(a[1], `"`)
		case "filename":
			h.Filename = strings.Trim(a[1], `"`)
		case "filename*":
			h.FilenameS = strings.Trim(a[1], `"`)
		default:
			return errors.New("invalid Content-Disposition directive")
		}
	}

	return nil
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
			if err := p.Unmarshal(headers[headerKey][0]); err != nil {
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

type Headers struct {
	ContentType        *ContentType       `header:"content-type"`
	ContentDisposition ContentDisposition `header:"content-disposition"`
	Server             string             `header:"server"`
	ContentLength      int                `header:"content-length"`
}

func main() {
	var headers Headers
	h := http.Header{
		"content-type":        []string{"application/json; charset=utf-8"},
		"content-disposition": []string{`form-data; name="fieldName"; filename="filename.jpg"`},
		"server":              []string{"Apache"},
		"content-length":      []string{"1234"},
	}
	err := Unmarshal(h, &headers)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", headers)
}
