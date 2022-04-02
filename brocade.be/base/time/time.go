package time

import (
	"fmt"
	"time"
)

//Definition
//$H[OROLOG]
//$Horolog gives date and time with one access.
//Its value is D , S where D is an integer value counting days since an origin specified below,
//and S is an integer value modulo 86,400 counting seconds.
//The value of $Horolog for the first second of December 31, 1840 is defined to be 0,0.
//S increases by 1 each second and S clears to 0 with a carry into D on the tick of midnight.

func Now() string {
	now := time.Now()
	return H(now)
}

func H(t time.Time) string {
	year := t.Year()
	month := t.Month()
	day := t.Day()
	h := t.Hour()
	m := t.Minute()
	s := t.Second()
	days := DaysUntil(year, int(month), day)
	rest := h*3600 + m*60 + s
	return fmt.Sprintf("%d,%d", days, rest)
}

func IsLeap(year int) bool {
	if year%4 != 0 {
		return false
	}
	if year%100 != 0 {
		return true
	}
	if year%400 == 0 {
		return true
	}
	return false
}

func DaysUntil(year, month, day int) int64 {
	var dm int64 = 0
	leap := IsLeap(year)
	if month > 1 {
		switch month - 1 {
		case 1:
			dm = 31
		case 2:
			dm = 59
		case 3:
			dm = 90
		case 4:
			dm = 120
		case 5:
			dm = 151
		case 6:
			dm = 181
		case 7:
			dm = 212
		case 8:
			dm = 243
		case 9:
			dm = 273
		case 10:
			dm = 304
		case 11:
			dm = 334
		case 12:
			dm = 365
		}
		if leap && month > 1 {
			dm += 1
		}
	}
	y := int64(year - 1)
	dy := y*365 + (y / 4) - (y / 100) + (y / 400) - 672046
	return dy + dm + int64(day)
}
