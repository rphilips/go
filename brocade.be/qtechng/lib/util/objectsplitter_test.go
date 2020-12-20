package util

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestIsObjStarter(t *testing.T) {
	type T struct {
		S string
		R bool
	}

	TestData := []T{
		{
			"",
			false,
		},
		{
			"Hello World",
			false,
		},
		{
			"m4_",
			false,
		},
		{
			"4_",
			false,
		},
		{
			"m4_A",
			true,
		},
		{
			"i4_A",
			true,
		},
		{
			"l4_A",
			false,
		},
		{
			"l4_N",
			false,
		},
		{
			"l4_Nphp",
			false,
		},
		{
			"l4_N_",
			false,
		},
		{
			"l4_Nphp_",
			false,
		},
		{
			"l4_N_a",
			true,
		},
		{
			"l4_Nphp_a",
			true,
		},
		{
			"l4_Nphp__",
			false,
		},
		{
			"r4_N",
			false,
		},
		{
			"r4_n",
			true,
		},
		{
			"r4__b",
			false,
		},
		{
			"m4_m4",
			true,
		},
		{
			"m4_i4",
			true,
		},
		{
			"m4_i4_a",
			false,
		},
		{
			"m4_i4_a",
			false,
		},
		{
			"m4_l4_a_aaa",
			true,
		},
	}

	for _, test := range TestData {
		blob := []byte(test.S)
		r := IsObjStarter(blob)

		if r != test.R {
			t.Errorf(fmt.Sprintf("\n\n%s\n\nfound: %v\nexpected: %v", string(blob), r, test.R))
			return
		}
	}
	return

}
func TestSplitter01(t *testing.T) {
	type T struct {
		S string
		R []string
	}

	TestData := []T{
		{
			"",
			[]string{""},
		},
		{
			"q",
			[]string{"q"},
		},
		{
			"Hello World",
			[]string{"Hello World"},
		},
		{
			"4_Hello",
			[]string{"4_Hello"},
		},
		{
			"4_",
			[]string{"4_"},
		},
		{
			"m4_",
			[]string{"m4_"},
		},
		{
			"m4_A",
			[]string{"", "m4_A", ""},
		},
		{
			"m4__",
			[]string{"m4__"},
		},
		{
			"m4_ABC(",
			[]string{"", "m4_ABC", "("},
		},
		{
			"Qm4_ABC(",
			[]string{"Q", "m4_ABC", "("},
		},
		{
			"m4_ABC(",
			[]string{"", "m4_ABC", "("},
		},
		{
			"m4_ABCm4_DEF",
			[]string{"", "m4_ABC", "", "m4_DEF", ""},
		},
		{
			"m4_ABCr4_ab_c_d_m4_DEF",
			[]string{"", "m4_ABC", "", "r4_ab_c_d_", "", "m4_DEF", ""},
		},
		{
			"Hellom4_ABCr4_ab_c_d_m4_DEF World",
			[]string{"Hello", "m4_ABC", "", "r4_ab_c_d_", "", "m4_DEF", " World"},
		},
		{

			"m4_ABCl4_Njs_H1there:World",
			[]string{"", "m4_ABC", "", "l4_Njs_H1there", ":World"},
		},
		{

			"m4_ABCl4_N_H1there:World",
			[]string{"", "m4_ABC", "", "l4_N_H1there", ":World"},
		},
		{

			"m4_m4",
			[]string{"", "m4_m4", ""},
		},
		{

			"m4_m4_m4",
			[]string{"m4_", "m4_m4", ""},
		},
		{

			"m4_m4_m4_m4",
			[]string{"", "m4_m4", "_", "m4_m4", ""},
		},
	}

	for _, test := range TestData {
		blob := []byte(test.S)
		r := ObjectSplitter(blob)
		rs := make([]string, 0)
		for _, part := range r {
			rs = append(rs, string(part))
		}

		rj, _ := json.MarshalIndent(rs, "", "    ")
		Rj, _ := json.MarshalIndent(test.R, "", "    ")

		srj := string(rj)
		sRj := string(Rj)

		if srj != sRj {
			t.Errorf(fmt.Sprintf("\n\n%s\n\nfound:\n%s\n\nexpected:\n%s", test.S, string(rj), string(Rj)))
			return
		}
	}
	return
}
