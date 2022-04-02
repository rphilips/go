package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	qfs "brocade.be/base/fs"
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

func MUPIP(args []string, cwd string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	qexe := "mupip"
	if cwd == "" {
		cwd, _ = os.Getwd()
	}
	pexe, e := exec.LookPath(qexe)
	if e != nil {
		return "", "", e
	}
	argums := []string{
		qexe,
	}
	argums = append(argums, args...)
	cmd := exec.Cmd{
		Path:   pexe,
		Args:   argums,
		Dir:    cwd,
		Stdout: &stdout,
		Stderr: &stderr,
	}

	err := cmd.Run()
	sout := stdout.String()
	serr := stderr.String()

	return sout, serr, err
}

func LoadDefaults() map[string]string {
	defaults := make(map[string]string)
	home, err := os.UserHomeDir()
	if err != nil {
		home, _ = os.Getwd()
	}
	config := filepath.Join(home, ".config", "goyo", "defaults.json")
	data, err := qfs.Fetch(config)
	if err != nil {
		json.Unmarshal(data, &defaults)
	}
	if err != nil {
		defaults["prompt-repl"] = "> "
	}
	return defaults
}

func KeyText(input string) (key string, text string) {
	input = strings.TrimSpace(input)
	k := strings.IndexAny(input, " \t")
	if k == -1 {
		return strings.ToLower(input), ""
	}
	return strings.ToLower(strings.TrimSpace(input[:k])), strings.TrimSpace(input[k+1:])
}
