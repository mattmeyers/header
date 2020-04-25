package header

type Marshaler interface {
	MarshalHeader() (string, error)
}

type Stringer interface {
	String() string
}
