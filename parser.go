package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
)

type Parser interface {
	Parse(string) error
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
	err := h.Parse(s)
	if err != nil {
		return nil, err
	}

	return h, nil
}

func (h *ContentType) Parse(s string) error {
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
	err := h.Parse(s)
	if err != nil {
		return nil, err
	}

	return h, nil
}

func (h *ContentDisposition) Parse(s string) error {
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

func Parse(headers http.Header, dst interface{}) error {
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("cannot parse")
	}

	rt := reflect.TypeOf(dst).Elem()
	for i := 0; i < rt.NumField(); i++ {
		fieldType := rt.FieldByIndex([]int{i})
		fieldVal := rv.Elem().FieldByIndex([]int{i})
		headerKey := fieldType.Tag.Get("header")
		switch headerKey {
		case "content-type":
			c, _ := NewContentType(headers[headerKey][0])
			fieldVal.Set(reflect.ValueOf(*c))
		case "content-disposition":
			c, _ := NewContentDisposition(headers[headerKey][0])
			fieldVal.Set(reflect.ValueOf(*c))
		case "server":
			c := NewServer(headers[headerKey][0])
			fieldVal.Set(reflect.ValueOf(*c))
		}
	}

	return nil
}

type Headers struct {
	ContentType        ContentType        `header:"content-type"`
	ContentDisposition ContentDisposition `header:"content-disposition"`
	Server             Server             `header:"server"`
}

func main() {
	var headers Headers
	h := http.Header{
		"content-type":        []string{"application/json; charset=utf-8"},
		"content-disposition": []string{`form-data; name="fieldName"; filename="filename.jpg"`},
		"server":              []string{"Apache"},
	}
	err := Parse(h, &headers)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", headers)
}
