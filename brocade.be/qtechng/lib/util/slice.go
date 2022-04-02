package util

import (
	"bytes"
	"sort"
	"strings"
)

//CmpByteSlice cheks if 2 slices contains the same data
func CmpByteSlice(first [][]byte, second [][]byte) bool {

	if first == nil {
		return second == nil
	}
	if second == nil {
		return first == nil
	}
	if len(first) != len(second) {
		return false
	}
	i := 0
	for i < len(first) {
		if !bytes.Equal(first[i], second[i]) {
			return false
		}
		i++
	}
	return true
}

// SliceToSet makes set from a slice
func SliceToSet(s []string) map[string]bool {
	m := make(map[string]bool)
	for _, k := range s {
		m[k] = true
	}
	return m
}

// DiffSlices neemt de vershillen m1-m2 en m2-m1 tussen 2 slices m1 en m2
func DiffSlices(m1, m2 []string) ([]string, []string) {
	sm1 := SliceToSet(m1)
	sm2 := SliceToSet(m2)
	x, y := DiffMaps(sm1, sm2)
	return x, y
}

// DiffMaps neemt de vershillen m1-m2 en m2-m1 tussen 2 maps m1 en m2
func DiffMaps(m1, m2 map[string]bool) (m1Notm2, m2Notm1 []string) {
	for k := range m1 {
		_, ok := m2[k]
		if !ok {
			m1Notm2 = append(m1Notm2, k)
		}
	}
	for k := range m2 {
		_, ok := m1[k]
		if !ok {
			m2Notm1 = append(m2Notm1, k)
		}
	}
	return
}

// CmpStringSlice vergelijkt 2 string slices (eventueel na sortering)
func CmpStringSlice(first []string, second []string, sorted bool) bool {

	if len(first) == 0 && len(second) == 0 {
		return true
	}
	if first == nil {
		return second == nil
	}
	if second == nil {
		return first == nil
	}
	if len(first) != len(second) {
		return false
	}
	if sorted {
		sort.Strings(first)
		sort.Strings(second)
	}
	i := 0
	for i < len(first) {
		if strings.Compare(first[i], second[i]) != 0 {
			return false
		}
		i++
	}
	return true
}
