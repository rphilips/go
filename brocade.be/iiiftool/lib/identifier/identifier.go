package identifier

type Identifier string

func (id Identifier) String() string {
	return string(id)
}

func (id Identifier) Location() string {
	location := "here"
	return location
}
