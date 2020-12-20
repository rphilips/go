package util

import "bytes"

// BMCreateTable maakt een Boyer-Moore tabel aan
func BMCreateTable(needle []byte) (bad [][]int, good []int, full []int) {
	lneedle := len(needle)
	if lneedle < 4 {
		return
	}
	// bad_character_table
	bad = make([][]int, 256)
	alfa := make([]int, 256)
	for i := 0; i < 256; i++ {
		bad[i] = []int{-1}
		alfa[i] = -1
	}

	for i, c := range needle {
		alfa[c] = i
		for j, a := range alfa {
			bad[j] = append(bad[j], a)
		}
	}

	// good_suffix_table

	good = make([]int, lneedle)
	for i := range good {
		good[i] = -1
	}

	rneedle := make([]byte, lneedle)
	for i, c := range needle {
		rneedle[lneedle-i-1] = c
	}

	n0 := fundamental(rneedle)
	n := make([]int, lneedle)
	for i, c := range n0 {
		n[lneedle-i-1] = c
	}
	for j := 0; j < lneedle-1; j++ {
		i := lneedle - n[j]
		if i != lneedle {
			good[i] = j
		}
	}

	// full shift tabel

	full = make([]int, lneedle)
	z0 := fundamental(needle)
	z := make([]int, lneedle)
	for i, c := range z0 {
		z[lneedle-i-1] = c
	}
	longest := 0
	for i, zv := range z {
		if zv == i+1 && zv > longest {
			longest = zv
		}
		full[lneedle-i-1] = longest
	}
	return
}

// BMSearch looks for the first index
func BMSearch(haystack []byte, needle []byte, bad [][]int, good []int, full []int) bool {
	lneedle := len(needle)
	lhaystack := len(haystack)
	if lneedle == 0 {
		return true
	}
	if lhaystack < lneedle {
		return false
	}
	if lneedle < 4 {
		return bytes.Index(haystack, needle) != -1
	}
	if lhaystack < 64 {
		return bytes.Index(haystack, needle) != -1
	}
	max := lhaystack
	if max > 1024 {
		max = 1024
	}
	if bytes.IndexByte(haystack[:max], 0) != -1 {
		return bytes.Index(haystack, needle) != -1
	}

	k := lneedle - 1
	previousK := -1

	for k < lhaystack {
		i := lneedle - 1
		h := k
		for i >= 0 && h > previousK && needle[i] == haystack[h] {
			i--
			h--
		}
		if i == -1 || h == previousK {
			return true
		}
		charShift := i - bad[haystack[h]][i]
		suffixShift := 0
		switch {
		case i+1 == lneedle:
			suffixShift = 1
		case good[i+1] == -1:
			suffixShift = lneedle - full[i+1]
		default:
			suffixShift = lneedle - 1 - good[i+1]
		}
		shift := charShift
		if suffixShift > shift {
			shift = suffixShift
		}
		if shift >= i+1 {
			previousK = k
		}
		k += shift
	}
	return false

}

func fundamental(needle []byte) (z []int) {
	lneedle := len(needle)

	// fundamental table
	z = make([]int, lneedle)
	if lneedle == 0 {
		return
	}
	if lneedle == 1 {
		z[0] = 1
		return
	}

	z[0] = lneedle
	z[1] = matchLength(needle, 0, 1)
	for i := 2; i <= z[1]; i++ {
		z[i] = z[1] - i + 1
	}
	l := 0
	r := 0
	for i := 2 + z[1]; i < lneedle; i++ {
		if i <= r {
			k := i - l
			b := z[k]
			a := r - i + 1
			if b < a {
				z[i] = b
			} else {
				z[i] = a + matchLength(needle, a, r+1)
				l = i
				r = i + z[i] - 1
			}
			continue
		}
		z[i] = matchLength(needle, 0, i)
		if z[i] > 0 {
			l = i
			r = i + z[i] - 1
		}
	}
	return
}

func matchLength(s []byte, idx1 int, idx2 int) int {
	ls := len(s)
	if idx1 == idx2 {
		return ls - idx1
	}
	count := 0
	for idx1 < ls && idx2 < ls && s[idx1] == s[idx2] {
		count++
		idx1++
		idx2++
	}
	return count
}
