package action

import (
	"fmt"
	"strings"

	qutil "brocade.be/goyo/lib/util"
	qyottadb "brocade.be/goyo/lib/yottadb"
	"github.com/eiannone/keyboard"
)

func walk(text string) []string {
	if text == "" {
		return nil
	}
	if err := keyboard.Open(); err != nil {
		panic(err)
	}
	defer func() {
		_ = keyboard.Close()
	}()
	gloref, _ := SplitRefValue(text)
	gloref2 := qyottadb.N(gloref)
	gloref1 := gloref
	gloref = gloref2
	stop := false

	for !stop {
		d, _ := qyottadb.D(gloref)
		if d < 1 || d == 10 {
			fmt.Println(gloref, fmt.Sprintf("$D=%d", d))
		} else {
			value, _ := qyottadb.G(gloref, true)
			fmt.Println(gloref+"="+value, fmt.Sprintf("$D=%d", d))
		}
		char, key, err := keyboard.GetKey()
		if err != nil {
			panic(err)
		}
		//fmt.Printf("You pressed: rune %q, key %X\r\n", char, key)
		if key == keyboard.KeyEsc || (char == 0 && (key == 3 || key == 4)) {
			stop = true
		}
		if char == 'h' {
			fmt.Println(`ESC: stop walking
d(own), n(ext), o(rder), ARROW-DOWN: $O(,+1)
u(p), p(prev), ARROW-UP            : $O(,-1)
r(ight), ARROW-RIGHT               : next level
l(eft), ARROW-LEFT                 : previous level
s(et)                              : set node
k(ill)                             : kill node
z(wr)                              : show node in ZWR
c(sv)                              : show node in CSV
e(dit)                             : edit node=value
/(search)                          : searches both on needle and regexp of needle`)
			gloref, _, err = qyottadb.Next(gloref)
			if err != nil {
				qutil.Error(err)
			}
			continue
		}
		if (char == 'd') || (char == 'n') || (char == 'o') || (char == 0 && fmt.Sprintf("%X", key) == "FFEC") {
			gloref, _, err = qyottadb.Next(gloref)
			if err != nil {
				qutil.Error(err)
			}
			continue
		}
		if (char == 'u') || (char == 'p') || (char == 0 && fmt.Sprintf("%X", key) == "FFED") {
			gloref, _, err = qyottadb.Prev(gloref)
			if err != nil {
				qutil.Error(err)
			}
			continue
		}
		if (char == 'r') || (char == 0 && fmt.Sprintf("%X", key) == "FFEA") {
			gloref, err = qyottadb.Right(gloref)
			if err != nil {
				qutil.Error(err)
			}
			continue
		}
		if (char == 'l') || (char == 0 && fmt.Sprintf("%X", key) == "FFEB") {
			gloref, err = qyottadb.Left(gloref)
			if err != nil {
				qutil.Error(err)
			}
			continue
		}
		if char == 's' {
			x := Set(gloref)
			if x == nil {
				continue
			}
			y := x[len(x)-1]
			y = strings.SplitN(y, " ", 2)[1]
			gloref, _ = SplitRefValue(y)
			continue
		}
		if char == 'k' {
			Kill(gloref, true)
			continue
		}
		if char == 'z' {
			ZWR(gloref)
			fmt.Println()
			continue
		}
		if char == 'c' {
			CSV(gloref, IOcsvout)
			fmt.Println()
			continue
		}
		if char == 'e' {
			x := Edit(gloref)
			if x != "" {
				gloref = x
			}
			continue
		}
		if char == '/' {
			needle := Ask("Search for: ", "")
			if needle == "" {
				continue
			}
			result := Search(gloref, needle, true)
			if result != "" {
				gloref = result
			}
			continue
		}
	}
	_ = keyboard.Close()
	h := []string{"walk " + gloref1}
	if gloref2 != gloref1 {
		h = append(h, "walk "+gloref2)
	}
	return h
}
