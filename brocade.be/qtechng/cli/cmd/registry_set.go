package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	qregistry "brocade.be/base/registry"
	"github.com/spf13/cobra"
)

var registrySetCmd = &cobra.Command{
	Use:   "set [key [value]]",
	Short: "Set registry keys",
	Long: `With two arguments, this command sets a registry key and value.
Without arguments, it interactively asks for the values`,
	Example: `
  qtechng registry set scratch-dir /home/rphilips/tmp
  qtechng registry set`,

	Args: cobra.MaximumNArgs(2),
	RunE: registrySet,
	Annotations: map[string]string{
		"remote-allowed":       "no",
		"rstrip-trailing-crlf": "yes",
	},
}

func init() {
	registryCmd.AddCommand(registrySetCmd)
}

func registrySet(cmd *cobra.Command, args []string) (err error) {
	setme := make(map[string]string)
	changed := false
	switch len(args) {
	case 2:
		setme[args[0]] = args[1]
		changed = true
	case 1:
		reader := bufio.NewReader(os.Stdin)
		key := args[0]
		value, ok := qregistry.Registry[key]
		if ok {
			fmt.Println(key, "->", value)
			fmt.Print("New value:")
		} else {
			fmt.Print("Value:")
		}
		text, _ := reader.ReadString('\n')
		value = strings.TrimSuffix(text, "\n")
		fmt.Println(key, "->", value)
		fmt.Print("S(et)/D(delete)/Q(uit): <S>")
		text, _ = reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if text == "" {
			text = "S"
		}
		text = strings.ToUpper(text)
		switch text {
		case "D":
			delete(qregistry.Registry, key)
			changed = true
		case "S":
			rvalue, ok := qregistry.Registry[key]
			if !ok || rvalue != value {
				setme[key] = value
				changed = true
			}
		}
	case 0:
		changed = askReg(setme)
	}

	if changed {
		for k, v := range setme {
			qregistry.SetRegistry(k, v)
			qregistry.Registry[k] = v
		}
	}

	return nil
}

func askReg(setme map[string]string) bool {
	changed := false
	reader := bufio.NewReader(os.Stdin)
	for i, item := range regmapqtechng {
		if i != 0 {
			fmt.Print("\n\n")
		}
		qt := item.qtechtype
		key := item.name
		ok := false
		if strings.Contains(qt, "B") && strings.Contains(qt, "P") && strings.Contains(qt, "W") {
			ok = true
		}
		ok = ok || strings.ContainsAny(qt, QtechType)
		if !ok {
			continue
		}
		mode := item.mode
		if mode == "skip" {
			continue
		}
		if mode == "set" {
			value := item.deffunc()
			rvalue, ok := qregistry.Registry[key]
			if !ok || rvalue != value {
				setme[key] = value
				changed = true
			}
			continue
		}
		rvalue, ok := qregistry.Registry[key]
		defa := rvalue
		if ok {
			fmt.Println(key, "->", rvalue)
			fmt.Println(item.doc)
			fmt.Print("New value:")
		} else {
			defa = item.deffunc()
			fmt.Println(key, "->", defa, "(default value)")
			fmt.Println(item.doc)
			fmt.Print("Value:")
		}
		text, _ := reader.ReadString('\n')
		value := strings.TrimSuffix(text, "\n")
		if value == "" {
			value = defa
		}
		fmt.Println(key, "->", value)
		fmt.Print("S(et)/D(delete)/N(ext): <S>")
		text, _ = reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if text == "" {
			text = "S"
		}
		text = strings.ToUpper(text)
		switch text {
		case "D":
			delete(qregistry.Registry, key)
		case "S":
			rvalue, ok := qregistry.Registry[key]
			if !ok || rvalue != value {
				setme[key] = value
				changed = true
			}
		}

	}
	return changed

}
