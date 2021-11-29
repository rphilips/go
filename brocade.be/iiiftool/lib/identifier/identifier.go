package identifier

type Identifier string

func (id Identifier) String() string {
	return string(id)
}
