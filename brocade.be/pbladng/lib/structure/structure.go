package structure

import (
	perror "brocade.be/pbladng/lib/error"
)

type Manifest struct {
	Letter string `json:"letter"`
	FName  string `json:"fname"`
	Digest string `json:"digest"`
}

type ImageID struct {
	Path   string `json:"path"`
	Mtime  string `json:"mtime"`
	Digest string `json:"digest"`
	Letter string `json:"letter"`
}

func (t *Topic) Test(hint string) (err error) {

	if t.Heading == "" {
		err = perror.Error("chapter-empty-title", t.Lineno, "chapter has no title [hint="+hint+"]")
		return
	}
	return
}
