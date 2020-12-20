package cmd

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func TestClip01(t *testing.T) {

	cmd := new(cobra.Command)
	clipboardSet(cmd, []string{"Hello World"})

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	clipboardGet(cmd, nil)
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = old

	if string(out) != "Hello World" {
		t.Errorf("Error: Found `%s`", out)
	}
}
