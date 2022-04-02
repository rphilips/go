package util

import (
	"fmt"

	qfs "brocade.be/base/fs"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
)

func Patch(fname1 string, fname2 string) (result string, err error) {

	btext1, err := qfs.Fetch(fname1)
	if err != nil {
		return
	}
	btext2, err := qfs.Fetch(fname2)
	if err != nil {
		return
	}
	edits := myers.ComputeEdits(span.URIFromPath(fname1), string(btext1), string(btext2))
	diff := fmt.Sprint(gotextdiff.ToUnified(fname1, fname2, string(btext1), edits))
	return diff, nil
}
