package holy

import (
	"sort"
	"strconv"
	"strings"
	"time"

	pregistry "brocade.be/pbladng/lib/registry"
	ptools "brocade.be/pbladng/lib/tools"
)

var loc = time.Now().Location()

var allholy []func(year int) (*time.Time, string, bool)

func init() {
	allholy = []func(year int) (*time.Time, string, bool){
		Pasen,
		Paasmaandag,
		Aswoensdag,
		Mariaopdracht,
		Advent1,
		Advent2,
		Advent3,
		Advent4,
		Kerstmis,
		Koningheelal,
		Heiligefamilie,
		Onschuldigekinderen,
		Silvester,
		Becket,
		Driekoningen,
		Hmariamoedergod,
		Doopjezus,
		Zondagvasten1,
		Zondagvasten2,
		Zondagvasten3,
		Zondagvasten4,
		Zondagvasten5,
		Mariaonbevlektontvangen,
		Hjozef,
		Mariaboodschap,
		Drievuldigheidszondag,
		Sacramentsdag,
		Lichtmis,
		Heilighartjezus,
		Johannesdedoper,
		Petruspaulus,
		Palmzondag,
		Belokenpasen,
		Pasen3,
		Pasen4,
		Pasen5,
		Pasen6,
		Pasen7,
		Stillezaterdag,
		Goedevrijdag,
		Wittedonderdag,
		Hemelvaart,
		Pinksteren,
		Mariahemelvaart,
		Allerheiligen,
		Allerzielen,
	}
}

func Today(today *time.Time) (result []string) {
	year := today.Year()
	for _, h := range allholy {
		t, note, _ := h(year)
		if !t.Equal(*today) {
			continue
		}
		result = append(result, note)
	}
	r := sundayinyear(today)
	if r != "" {
		result = append(result, r)
	}

	sort.Strings(result)
	return

}

func sundayinyear(today *time.Time) string {
	if !issunday(today) {
		return ""
	}
	year := today.Year()
	scor := pregistry.Registry["sunday-correction"].(string)
	icor, _ := strconv.Atoi(scor)
	psun, _, _ := Doopjezus(year)
	add := 1 + icor
	sun := *psun
	if sun.Equal(*today) {
		return strconv.Itoa(add) + "e ZONDAG DOOR HET JAAR"
	}
	aswo, _, _ := Aswoensdag(year)
	pink, _, _ := Pinksteren(year)
	adv1, _, _ := Advent1(year)
	for {
		sun = sun.AddDate(0, 0, 7)
		if sun.Before(*aswo) {
			add++
			if sun.Equal(*today) {
				return strconv.Itoa(add) + "e ZONDAG DOOR HET JAAR"
			}
		} else {
			break
		}
	}
	sun = *pink
	sun = sun.AddDate(0, 0, -14)

	for {
		sun = sun.AddDate(0, 0, 7)
		if sun.Before(*adv1) {
			add++
			if sun.Equal(*today) {
				return strconv.Itoa(add) + "e ZONDAG DOOR HET JAAR"
			}
		} else {
			break
		}
	}
	return ""

}

func myday(y, m, d int) *time.Time {
	x := time.Date(y, time.Month(m), d, 0, 0, 0, 0, loc)
	return &x
}

func addto(t *time.Time, days int) *time.Time {
	day := t.AddDate(0, 0, days)
	return &day
}

func issunday(t *time.Time) bool {
	return t.Weekday() == time.Sunday
}

func Pasen(year int) (*time.Time, string, bool) {
	g := year % 19
	e := 0
	c := year / 100
	h := (c - c/4 - (8*c+13)/25 + 19*g + 15) % 30
	i := h - (h/28)*(1-(h/28)*(29/(h+1))*((21-g)/11))
	j := (year + year/4 + i + 2 - c + c/4) % 7
	p := i - j + e
	d := 1 + (p+27+(p+6)/40)%31
	m := 3 + (p+26)/30
	t := time.Date(year, time.Month(m), d, 0, 0, 0, 0, loc)
	return &t, "*PASEN - VERRIJZENISZONDAG*", true
}

func Paasmaandag(year int) (*time.Time, string, bool) {
	eastern, _, _ := Pasen(year)
	monday := eastern.AddDate(0, 0, 1)
	return &monday, "PAASMAANDAG", false
}

func Aswoensdag(year int) (*time.Time, string, bool) {
	eastern, _, _ := Pasen(year)
	monday := eastern.AddDate(0, 0, -46)
	return &monday, "*ASWOENSDAG - Begin van de Veertigdagentijd*", false
}

func Mariaopdracht(year int) (*time.Time, string, bool) {
	t := time.Date(year, 11, 21, 0, 0, 0, 0, loc)
	return &t, "*Maria-Opdracht*", false
}

func Advent1(year int) (*time.Time, string, bool) {
	day := time.Date(year, 11, 27, 0, 0, 0, 0, loc)
	for {
		weekday := ptools.StringDate(&day, "D")
		if strings.HasPrefix(weekday, "zondag") {
			return &day, "*1e ZONDAG VAN DE ADVENT*", false
		}
		day = day.AddDate(0, 0, 1)
	}
}

func Advent2(year int) (*time.Time, string, bool) {
	day, _, _ := Advent1(year)
	xday := day.AddDate(0, 0, 7)
	return &xday, "*2e ZONDAG VAN DE ADVENT*", false
}

func Advent3(year int) (*time.Time, string, bool) {
	day, _, _ := Advent1(year)
	xday := day.AddDate(0, 0, 14)
	return &xday, "*3e ZONDAG VAN DE ADVENT*", false
}

func Advent4(year int) (*time.Time, string, bool) {
	day, _, _ := Advent1(year)
	xday := day.AddDate(0, 0, 21)
	return &xday, "*4e ZONDAG VAN DE ADVENT*", false
}

func Kerstmis(year int) (*time.Time, string, bool) {
	day := time.Date(year, 12, 25, 0, 0, 0, 0, loc)
	return &day, "*KERSTMIS - GEBOORTE VAN DE HEER*", true
}

func Koningheelal(year int) (*time.Time, string, bool) {
	day, _, _ := Advent1(year)
	xday := day.AddDate(0, 0, -21)
	return &xday, "*CHRISTUS KONING VAN HET HEELAL*", false
}

func Heiligefamilie(year int) (*time.Time, string, bool) {
	datum := myday(year, 12, 30)
	start := myday(year, 12, 26)
	end := myday(year, 12, 26)
	for {
		if issunday(start) {
			datum = start
			break
		}
		xstart := start.AddDate(0, 0, 1)
		if end.Before(xstart) {
			break
		}
		start = &xstart
	}
	return datum, "FEEST VAN DE HEILIGE FAMILIE: Jezus, Maria, Jozef", false
}

func Onschuldigekinderen(year int) (*time.Time, string, bool) {
	return myday(year, 12, 28), "HH. ONSCHULDIGE KINDEREN", false
}

func Silvester(year int) (*time.Time, string, bool) {
	return myday(year, 12, 31), "*H. Silvester I,* paus", false
}

func Becket(year int) (*time.Time, string, bool) {
	return myday(year, 12, 29), "*H. Thomas Becket,* bisschop en martelaar", false
}

func Driekoningen(year int) (*time.Time, string, bool) {
	return myday(year, 1, 6), "*OPENBARING VAN DE HEER* (Driekoningen)", false
}

func Hmariamoedergod(year int) (*time.Time, string, bool) {
	return myday(year, 1, 1), "*H. MARIA, MOEDER VAN GOD*", false
}

func Doopjezus(year int) (*time.Time, string, bool) {
	king3, _, _ := Driekoningen(year)
	if !issunday(king3) {
		king3 = myday(year, 1, 2)
		for {
			if issunday(king3) {
				break
			}
			xking3 := king3.AddDate(0, 0, 1)
			king3 = &xking3
		}
	}

	datum := king3.AddDate(0, 0, 7)
	return &datum, "*DOOPSEL VAN CHRISTUS*", false
}

func Zondagvasten1(year int) (*time.Time, string, bool) {
	datum, _, _ := Aswoensdag(year)
	xdatum := datum.AddDate(0, 0, 4)
	return &xdatum, "1e ZONDAG IN DE VEERTIGDAGENTIJD", false
}

func Zondagvasten2(year int) (*time.Time, string, bool) {
	datum, _, _ := Aswoensdag(year)
	xdatum := datum.AddDate(0, 0, 11)
	return &xdatum, "2e ZONDAG IN DE VEERTIGDAGENTIJD", false
}

func Zondagvasten3(year int) (*time.Time, string, bool) {
	datum, _, _ := Aswoensdag(year)
	xdatum := datum.AddDate(0, 0, 18)
	return &xdatum, "3e ZONDAG IN DE VEERTIGDAGENTIJD", false
}

func Zondagvasten4(year int) (*time.Time, string, bool) {
	datum, _, _ := Aswoensdag(year)
	xdatum := datum.AddDate(0, 0, 25)
	return &xdatum, "4e ZONDAG IN DE VEERTIGDAGENTIJD", false
}

func Zondagvasten5(year int) (*time.Time, string, bool) {
	datum, _, _ := Aswoensdag(year)
	xdatum := datum.AddDate(0, 0, 32)
	return &xdatum, "5e ZONDAG IN DE VEERTIGDAGENTIJD", false
}

func Mariaonbevlektontvangen(year int) (*time.Time, string, bool) {
	return myday(year, 12, 8), "*MARIA ONBEVLEKT ONTVANGEN*", false
}

func Hjozef(year int) (*time.Time, string, bool) {
	return myday(year, 3, 19), "*H. Jozef*", false
}

func Mariaboodschap(year int) (*time.Time, string, bool) {
	return myday(year, 3, 25), "*Aankondiging van de Heer (Maria Boodschap)*", false
}

func Drievuldigheidszondag(year int) (*time.Time, string, bool) {
	p, _, _ := Pinksteren(year)
	return addto(p, 7), "*HOOGFEEST VAN DE DRIE-EENHEID*", false
}

func Sacramentsdag(year int) (*time.Time, string, bool) {
	p, _, _ := Pasen(year)
	return addto(p, 60), "*SACRAMENTSDAG: H. LICHAAM EN BLOED VAN CHRISTUS*", false
}

func Lichtmis(year int) (*time.Time, string, bool) {
	return myday(year, 2, 2), "OPDRACHT VAN DE HEER (MARIA-LICHTMIS)", false
}

func Heilighartjezus(year int) (*time.Time, string, bool) {
	p, _, _ := Pasen(year)
	return addto(p, 68), "HEILIG HART VAN JEZUS", false
}

func Johannesdedoper(year int) (*time.Time, string, bool) {
	return myday(year, 6, 24), "*GEBOORTE VAN DE HEILIGE JOHANNES DE DOPER*", false
}

func Petruspaulus(year int) (*time.Time, string, bool) {
	return myday(year, 6, 29), "Hoogfeest van de *Heilige Petrus en Paulus,* apostelen", false
}

func Palmzondag(year int) (*time.Time, string, bool) {
	p, _, _ := Pasen(year)
	return addto(p, -7), "*PALMZONDAG - PASSIE VAN DE HEER*", false
}

func Belokenpasen(year int) (*time.Time, string, bool) {
	p, _, _ := Pasen(year)
	return addto(p, 7), "2de ZONDAG VAN PASEN (Beloken Pasen)", false
}

func Pasen3(year int) (*time.Time, string, bool) {
	p, _, _ := Pasen(year)
	return addto(p, 14), "3e ZONDAG VAN PASEN", false
}

func Pasen4(year int) (*time.Time, string, bool) {
	p, _, _ := Pasen(year)
	return addto(p, 21), "4e ZONDAG VAN PASEN", false
}

func Pasen5(year int) (*time.Time, string, bool) {
	p, _, _ := Pasen(year)
	return addto(p, 28), "5e ZONDAG VAN PASEN", false
}

func Pasen6(year int) (*time.Time, string, bool) {
	p, _, _ := Pasen(year)
	return addto(p, 35), "6e ZONDAG VAN PASEN", false
}

func Pasen7(year int) (*time.Time, string, bool) {
	p, _, _ := Pasen(year)
	return addto(p, 42), "7e ZONDAG VAN PASEN", false
}

func Stillezaterdag(year int) (*time.Time, string, bool) {
	p, _, _ := Pasen(year)
	return addto(p, -1), "STILLE ZATERDAG", false
}

func Goedevrijdag(year int) (*time.Time, string, bool) {
	p, _, _ := Pasen(year)
	return addto(p, -2), "GOEDE VRIJDAG", false
}

func Wittedonderdag(year int) (*time.Time, string, bool) {
	p, _, _ := Pasen(year)
	return addto(p, -3), "WITTE DONDERDAG", false
}

func Hemelvaart(year int) (*time.Time, string, bool) {
	p, _, _ := Pasen(year)
	return addto(p, 39), "*ONS-HEERHEMELVAART*", false
}

func Pinksteren(year int) (*time.Time, string, bool) {
	p, _, _ := Pasen(year)
	return addto(p, 49), "*PINKSTEREN*", true
}

func Mariahemelvaart(year int) (*time.Time, string, bool) {
	return myday(year, 8, 15), "*TENHEMELOPNEMING VAN MARIA*", true
}

func Allerheiligen(year int) (*time.Time, string, bool) {
	return myday(year, 11, 1), "*ALLERHEILIGEN*", true
}

func Allerzielen(year int) (*time.Time, string, bool) {
	return myday(year, 11, 2), "*ALLERZIELEN*", false
}
