package header

import "strconv"

type String string

func (h *String) UnmarshalHeader(s string) error { *h = String(s); return nil }

func (h String) String() string { return string(h) }

type Int int

func (h *Int) UnmarshalHeader(s string) error {
	i, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*h = Int(i)
	return nil
}

func (h Int) String() string { return strconv.Itoa(int(h)) }
