package identifier

import (
	"encoding/base64"
	"log"
)

type Identifier string

func (id Identifier) String() string {
	return string(id)
}

func (id Identifier) Location() string {
	location := "here"
	return location
}

// Encode in base64url
func (id Identifier) Encode() string {
	return base64.URLEncoding.EncodeToString([]byte(id))
}

// Decode from base64url
func (id Identifier) Decode() string {
	dec, err := base64.URLEncoding.DecodeString(id.String())
	if err != nil {
		log.Fatal("error decoding string:\n", err)
	}
	return string(dec)
}
