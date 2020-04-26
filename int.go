package header

import "strconv"

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

type ContentLength struct{ Int }

func NewContentLength(i int) ContentLength { return ContentLength{Int(i)} }
