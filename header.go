package header

import (
	"errors"
	"strings"
)

type StringHeader string

// func (h *StringHeader) Unmarshal(s string) error {
// 	h.Value = s
// 	return nil
// }

type Server string

func NewServer(s string) Server { return Server(s) }

func (h *Server) UnmarshalHeader(s string) error { *h = Server(s); return nil }

func (h Server) String() string { return string(h) }

func (h Server) MarshalHeader() (string, error) { return h.String(), nil }

type ContentType struct {
	MediaType string
	Charset   string
	Boundary  string
}

func NewContentType(s string) (ContentType, error) {
	h := ContentType{}
	err := h.UnmarshalHeader(s)
	if err != nil {
		return ContentType{}, err
	}

	return h, nil
}

func (h *ContentType) UnmarshalHeader(s string) error {
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

func (h ContentType) String() string {
	var sb strings.Builder
	sb.WriteString(h.MediaType)
	if h.Charset != "" {
		sb.WriteString("; charset=")
		sb.WriteString(h.Charset)
	}
	if h.Boundary != "" {
		sb.WriteString("; boundary=")
		sb.WriteString(h.Boundary)
	}
	return sb.String()
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
	err := h.UnmarshalHeader(s)
	if err != nil {
		return nil, err
	}

	return h, nil
}

func (h *ContentDisposition) UnmarshalHeader(s string) error {
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
