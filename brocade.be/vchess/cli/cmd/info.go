package cmd

import (
	"encoding/json"
	"sort"
	"strings"
	"time"

	vicyear "brocade.be/vchess/lib/icyear"
	vregistry "brocade.be/vchess/lib/registry"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Information info `vchess`",
	Long:  `Version and build time information info the vchess executable`,

	Args:    cobra.NoArgs,
	Example: `vchess info`,
	RunE:    info,
}

func init() {

	rootCmd.AddCommand(infoCmd)
}

func info(cmd *cobra.Command, args []string) error {

	msg := make(map[string]string)

	for _, s := range []string{"kbsb", "club"} {
		m := vregistry.Registry[s].(map[string]any)
		for k, v := range m {
			msg[s+"-"+k] = v.(string)
		}

	}
	teams := vicyear.Teams(nil, nil)
	tm := make([]string, 0, len(teams))
	for _, team := range teams {
		k := team.Name
		x := make([]string, 0)
		x = append(x, team.Nr)
		x = append(x, team.Division)
		x = append(x, team.PK)
		v := strings.Join(x, "/")
		tm = append(tm, k+"["+v+"]")
	}
	sort.Strings(tm)
	msg["ic-teams"] = strings.Join(tm, ", ")

	b, _ := json.MarshalIndent(msg, "", "    ")
	Fmsg = string(b)

	matches := vicyear.Matches(nil, nil, "")

	start := true
	after := ""
	for _, match := range matches {
		round := match.Round
		date := match.Date
		home := match.Home
		remote := match.Remote
		if start {
			start = false
			after = round + " (" + date.Format(time.RFC3339)[:10] + ")"
			Fmsg += "\n\n" + after + ":\n"
		}
		Fmsg += "    " + home.Division + ": " + home.Name + " - " + remote.Name + "\n"

	}
	return nil
}
