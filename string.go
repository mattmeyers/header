package header

type String string

func (h *String) UnmarshalHeader(s string) error { *h = String(s); return nil }

func (h String) String() string { return string(h) }

type Server struct{ String }

func NewServer(s string) Server { return Server{String(s)} }
