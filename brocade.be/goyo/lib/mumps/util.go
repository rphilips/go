package mumps

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var Number1 = regexp.MustCompile(`^[+-]?(([0-9]+(\.[0-9]+)?)|(\.[0-9]+))(E[+-]?[0-9]+)$`)
var Number2 = regexp.MustCompile(`^[+-]?[0-9]+$`)
var Number3 = regexp.MustCompile(`^[+-]?[0-9]+E[+]?[0-9]+$`)

func Escape(s string) (r string) {
	r = strings.ReplaceAll(s, "\\\\", "\x00")
	r = strings.ReplaceAll(r, "\\/", "\x01")
	r = strings.ReplaceAll(r, "\\=", "\x02")
	return r
}

func Unescape(s string) (r string) {
	r = strings.ReplaceAll(s, "\x02", "\\=")
	r = strings.ReplaceAll(r, "\x01", "\\/")
	r = strings.ReplaceAll(r, "\x00", "\\\\")
	return r
}

func Nature(s string) (string, string, error) {

	switch {
	case Number2.MatchString(s):
		sign := 1
		if strings.HasPrefix(s, "-") {
			sign = -1
		}
		s = strings.TrimLeft(s, "0+-")
		if s == "" {
			return "0", "i", nil
		}
		z, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return s, "s", err
		}
		return strconv.FormatInt(int64(sign)*z, 10), "i", nil

	case Number3.MatchString(s):
		x := strings.SplitN(s, "E", -1)
		y := strings.ReplaceAll(x[len(x)-1], "+", "")
		z, err := strconv.ParseInt(y, 10, 0)
		if err != nil {
			return s, "s", err
		}
		if z == 0 {
			return Nature(x[0])
		}
		if z > 64 {
			return s, "s", errors.New("exponent too big")
		}
		ext := strings.Repeat("0", int(z))
		return Nature(x[0] + ext)

	case Number1.MatchString(s):
		x, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return strconv.FormatFloat(x, 'G', -1, 64), "f", nil
		}
		return s, "s", err

	}
	return s, "s", nil
}
