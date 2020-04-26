package header

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Age is the number of seconds an object has been in a proxy cache.
// This value must be a nonnegative integer.
//
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Age
type Age struct{ Int }

// NewAge creates an Age header. This value cannot be negative.
func NewAge(i int) (Age, error) {
	if i < 0 {
		return Age{}, errors.New("age cannot be negative")
	}
	return Age{Int(i)}, nil
}

type TimeDelta struct {
	Name  string
	Value int
	Valid bool
}

func (t TimeDelta) String() string {
	var sb strings.Builder
	sb.WriteString(t.Name)
	if t.Valid {
		sb.WriteString("=")
		sb.WriteString(strconv.Itoa(t.Value))
	}
	return sb.String()
}

type Directive struct {
	Name  string
	Value NullString
}

func (d Directive) String() string {
	var sb strings.Builder
	sb.WriteString(d.Name)
	if d.Value.Valid {
		sb.WriteString("=")
		sb.WriteString(d.Value.String)
	}
	return sb.String()
}

// CacheControl holds directives for response and request caching. All time
// deltas are measured in seconds. If a time delta directive is not set, the
// struct field will be nil. Any other directives will be false if not
// provided.
//
// Refer to RFC 7234, section 5.2 for the formal specification.
//
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control
type CacheControl struct {
	// Cacheability
	Public  bool
	Private bool
	NoCache bool
	NoStore bool
	// Expiration
	MaxAge   *TimeDelta
	SMaxAge  *TimeDelta
	MaxStale *TimeDelta
	MinFresh *TimeDelta
	// Revalidation and reloading
	MustRevalidate  bool
	ProxyRevalidate bool
	Immutable       bool
	// Other
	NoTransform  bool
	OnlyIfCached bool
	Extensions   []Directive
}

func (h *CacheControl) UnmarshalHeader(s string) error {
	if s == "" {
		return nil
	}

	h.Extensions = make([]Directive, 0)

	directives := strings.Split(s, ",")
	for _, d := range directives {
		switch i := strings.TrimSpace(d); {
		case i == "public":
			h.Public = true
		case i == "private":
			h.Private = true
		case i == "no-cache":
			h.NoCache = true
		case i == "no-store":
			h.NoStore = true
		case strings.HasPrefix(i, "max-age"):
			v, err := parseTimeDelta(i, true)
			if err != nil {
				return fmt.Errorf("invalid max-age directive: %v", err)
			}
			h.MaxAge = v
		case strings.HasPrefix(i, "s-maxage"):
			v, err := parseTimeDelta(i, true)
			if err != nil {
				return fmt.Errorf("invalid s-maxage directive: %v", err)
			}
			h.SMaxAge = v
		case strings.HasPrefix(i, "max-stale"):
			v, err := parseTimeDelta(i, false)
			if err != nil {
				return fmt.Errorf("invalid max-stale directive: %v", err)
			}
			h.MaxStale = v
		case strings.HasPrefix(i, "min-fresh"):
			v, err := parseTimeDelta(i, true)
			if err != nil {
				return fmt.Errorf("invalid min-fresh directive: %v", err)
			}
			h.MinFresh = v
		case i == "must-revalidate":
			h.MustRevalidate = true
		case i == "proxy-revalidate":
			h.ProxyRevalidate = true
		case i == "immutable":
			h.Immutable = true
		case i == "no-transform":
			h.NoTransform = true
		case i == "only-if-cached":
			h.OnlyIfCached = true
		default:
			e := strings.Index(i, "=")
			var newD Directive
			if e < 0 {
				newD.Name = i
			} else {
				newD.Name = i[:e]
				newD.Value = NullString{String: i[e+1:], Valid: true}
			}
			h.Extensions = append(h.Extensions, newD)
		}
	}

	return nil
}

func parseTimeDelta(s string, required bool) (*TimeDelta, error) {
	e := strings.Index(s, "=")
	if e < 0 {
		if required {
			return nil, errors.New("no \"=\" present")
		} else {
			return &TimeDelta{Name: s, Valid: false}, nil
		}
	}
	v, err := strconv.Atoi(s[e+1:])
	if err != nil || v < 0 {
		return nil, errors.New("invalid time delta directive")
	}
	return &TimeDelta{Name: s[:e], Value: v, Valid: true}, nil
}

func (h CacheControl) String() string {
	directives := make([]string, 0)
	if h.Public {
		directives = append(directives, "public")
	}
	if h.Private {
		directives = append(directives, "private")
	}
	if h.NoCache {
		directives = append(directives, "no-cache")
	}
	if h.NoStore {
		directives = append(directives, "no-store")
	}
	if h.MaxAge != nil {
		directives = append(directives, h.MaxAge.String())
	}
	if h.SMaxAge != nil {
		directives = append(directives, h.SMaxAge.String())
	}
	if h.MaxStale != nil {
		directives = append(directives, h.MaxStale.String())
	}
	if h.MinFresh != nil {
		directives = append(directives, h.MinFresh.String())
	}
	if h.MustRevalidate {
		directives = append(directives, "must-revalidate")
	}
	if h.ProxyRevalidate {
		directives = append(directives, "proxy-revalidate")
	}
	if h.Immutable {
		directives = append(directives, "immutable")
	}
	if h.NoTransform {
		directives = append(directives, "no-transform")
	}
	if h.OnlyIfCached {
		directives = append(directives, "only-if-cached")
	}

	for _, e := range h.Extensions {
		directives = append(directives, e.String())
	}
	return strings.Join(directives, ", ")
}
