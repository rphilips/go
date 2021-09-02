package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/abiosoft/ishell/v2"
	"github.com/eiannone/keyboard"

	qmumps "brocade.be/goyo/lib/mumps"
)

func walk(c *ishell.Context) {
	stop := false
	gloref := ""
	if len(c.RawArgs) > 1 {
		gloref = strings.Join(c.RawArgs[1:], " ")
		gloref, _, _ = qmumps.EditGlobal(gloref)
	}

	if gloref == "" && Fgloref != "" {
		gloref = Fgloref
	}
	c.ShowPrompt(false)
	if gloref == "" {
		c.ShowPrompt(true)
		return
	}
	c.Println(qmumps.Ask("Esc|up|down|left|right|s|k|K"))
	if err := keyboard.Open(); err != nil {
		panic(err)
	}
	defer func() {
		_ = keyboard.Close()
	}()
	for !stop {
		value, _ := qmumps.GlobalValue(gloref)
		d := qmumps.GlobalDefined(gloref)
		if d != 1 && d != 11 {
			c.Println(qmumps.Info(gloref), " ", qmumps.Error("$D()="+strconv.Itoa(d)))
		} else {
			c.Println(qmumps.Info(gloref)+"="+qmumps.Info(value), qmumps.Error("$D()="+strconv.Itoa(d)))
		}
		char, key, err := keyboard.GetKey()
		if err != nil {
			panic(err)
		}
		fmt.Printf("You pressed: rune %q, key %X\r\n", char, key)
		if key == keyboard.KeyEsc {
			stop = true
		}
		if (char == 'd') || (char == 'n') || (char == 0 && fmt.Sprintf("%X", key) == "FFEC") {
			gloref, _ = qmumps.GlobalNext(gloref)
			continue
		}
		if (char == 'u') || (char == 'p') || (char == 0 && fmt.Sprintf("%X", key) == "FFED") {
			gloref, _ = qmumps.GlobalPrev(gloref)
			continue
		}
		if (char == 'r') || (char == 0 && fmt.Sprintf("%X", key) == "FFEA") {
			gloref, _ = qmumps.GlobalRight(gloref)
			continue
		}
		if (char == 'l') || (char == 0 && fmt.Sprintf("%X", key) == "FFEB") {
			gloref, _ = qmumps.GlobalLeft(gloref)
			continue
		}
		if char == 's' {
			_ = keyboard.Close()
			c.Println(qmumps.Ask("ref=value (empty to quit):"))
			gloref, _, _ := handleset(c, gloref)
			if gloref != "" {
				Fgloref = gloref
			}
			//c.Println(qmumps.Info(gloref) + "=" + qmumps.Info(value) + " " + qmumps.Error("$D()="+strconv.Itoa(d)))
			if err := keyboard.Open(); err != nil {
				panic(err)
			}
			defer func() {
				_ = keyboard.Close()
			}()
			continue
		}
		if char == 'k' {
			qmumps.GlobalKillk(gloref)
			continue
		}
		if char == 'K' {
			qmumps.GlobalKillK(gloref)
			continue
		}

	}
	_ = keyboard.Close()
	c.ShowPrompt(true)
}
