package cmd

import (
	"embed"
	"html/template"

	"github.com/spf13/cobra"
)

//go:embed templates
var guifs embed.FS

// Finplace replace the file contents

var guiCmd = &cobra.Command{
	Use:     "gui",
	Short:   "GUI functions",
	Long:    `All kinds of GUI functions`,
	Args:    cobra.NoArgs,
	Example: "qtechng gui",
}

type GuiFiller struct {
	CSS   template.CSS
	JS    template.JS
	Title string
	Vars  map[string]string
}

var guiFiller GuiFiller

func init() {
	rootCmd.AddCommand(guiCmd)
	b, _ := guifs.ReadFile("templates/gui.css")
	guiFiller.CSS = template.CSS(b)
	b, _ = guifs.ReadFile("templates/gui.js")
	guiFiller.JS = template.JS(b)
	guiFiller.Title = "QtechNG"
}
