package util

import (
	"strings"
	"testing"
)

func TestStrReverse(t *testing.T) {
	test := " !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"
	expected := "~}|{zyxwvutsrqponmlkjihgfedcba`_^]\\[ZYXWVUTSRQPONMLKJIHGFEDCBA@?>=<;:9876543210/.-,+*)('&%$#\"! "
	result := StrReverse(test)
	Check(result, expected, t)
}

func TestURLSafe(t *testing.T) {
	test := "HELLO !\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~ WORLD"
	expected := "hello__________________________________world"
	result := URLSafe(test)
	Check(result, expected, t)
}

func TestGmConvertArgs(t *testing.T) {
	expected := []string{
		"convert", "-flatten", "-quality", "1",
		"-define", "jp2:prg=rlcp", "-define", "jp2:numrlvls=7",
		"-define", "jp2:tilewidth=2",
		"-define", "jp2:tileheight=2"}
	result := GmConvertArgs(1, 2)
	Check(strings.Join(result, ""), strings.Join(expected, ""), t)
}
